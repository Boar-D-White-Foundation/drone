package dbq

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/boar-d-white-foundation/drone/db"
)

func queueKey(name string) string {
	return "dbq:queue:" + name
}

func queueDLXKey(name string) string {
	return queueKey(name) + ":dlx"
}

type handler interface {
	do(ctx context.Context, tx db.Tx, task any) (int, any, error)
	getQueue(tx db.Tx, key string) ([]any, error)
	setQueue(tx db.Tx, key string, queue []any) error
}

type Registry struct {
	queue    *Queue
	handlers map[string]handler
}

func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[string]handler),
	}
}

type Queue struct {
	registry     *Registry
	database     db.DB
	taskEnqueued chan struct{}
}

func NewQueue(
	registry *Registry,
	database db.DB,
) (*Queue, error) {
	if registry.queue != nil {
		return nil, errors.New("registrty is already bound to another queue")
	}

	result := Queue{
		registry: registry,
		database: database,
		// allow burst for 25 tasks, we can miss an added task if there's another executing
		// in such case we'll wait for pollDelay to consume it instead of consuming it immediately
		taskEnqueued: make(chan struct{}, 25),
	}
	registry.queue = &result
	return &result, nil
}

func (q *Queue) StartHandlers(ctx context.Context, pollDelay time.Duration) {
	for {
		err := q.database.Do(ctx, func(tx db.Tx) error {
			// consume just 1 task to release db lock fast
			// pick regular tasks before dlx
			for k, handler := range q.registry.handlers {
				key, dlxKey := queueKey(k), queueDLXKey(k)
				queue, err := handler.getQueue(tx, key)
				if err != nil {
					return err
				}
				dlx, err := handler.getQueue(tx, dlxKey)
				if err != nil {
					return err
				}

				var task any
				var selected []any
				var selectedKey string
				switch {
				case len(queue) > 0:
					task, queue = queue[0], queue[1:]
					selected, selectedKey = queue, key
				case len(dlx) > 0:
					task, dlx = dlx[0], dlx[1:]
					selected, selectedKey = dlx, dlxKey
				default:
					continue
				}

				slog.Info("fetched task to execute", slog.Any("task", task))
				if err := handler.setQueue(tx, selectedKey, selected); err != nil {
					return err
				}

				if ttl, task, err := handler.do(ctx, tx, task); err != nil {
					// move task to dlx to deal with it later
					logCtx := []any{slog.Any("err", err), slog.String("key", key), slog.Any("task", task)}
					if ttl < 1 {
						slog.Error("err executing task, ttl is zero, stopping retrying", logCtx...)
						break
					}

					slog.Error("err executing task, ttl is not zero, moving to dlx", logCtx...)
					dlx = append(dlx, task)
					if err := handler.setQueue(tx, dlxKey, dlx); err != nil {
						return err
					}
					break
				}

				slog.Info("finished executing task", slog.Any("task", task))
				break
			}

			return nil
		})
		if err != nil {
			slog.Error("err consume task", slog.Any("err", err))
			continue
		}

		select {
		case <-time.After(pollDelay):
		case <-q.taskEnqueued:
		case <-ctx.Done():
			return
		}
	}
}

type Handler[T any] func(context.Context, db.Tx, T) error

func (h Handler[T]) do(ctx context.Context, tx db.Tx, task any) (int, any, error) {
	casted, ok := task.(dbTask[T])
	if !ok {
		return 0, nil, fmt.Errorf("invalid task type %T: %+v", task, task)
	}

	casted.TTL--
	return casted.TTL, casted, h(ctx, tx, casted.Args)
}

func (h Handler[T]) getQueue(tx db.Tx, key string) ([]any, error) {
	queue, err := db.GetJsonDefault[[]dbTask[T]](tx, key, nil)
	if err != nil {
		return nil, fmt.Errorf("get queue %q: %w", key, err)
	}

	result := make([]any, 0, len(queue))
	for _, task := range queue {
		result = append(result, task)
	}
	return result, nil
}

func (h Handler[T]) setQueue(tx db.Tx, key string, queue []any) error {
	castedQueue := make([]dbTask[T], 0, len(queue))
	for _, task := range queue {
		castedTask, ok := task.(dbTask[T])
		if !ok {
			return fmt.Errorf("invalid task type %T: %+v", task, task)
		}

		castedQueue = append(castedQueue, castedTask)
	}

	if err := db.SetJson(tx, key, castedQueue); err != nil {
		return fmt.Errorf("set queue %q: %w", key, err)
	}

	return nil
}

type dbTask[T any] struct {
	Name string `json:"name"`
	TTL  int    `json:"ttl"`
	Args T      `json:"args"`
}

type Task[T any] struct {
	name     string
	registry *Registry
}

func (t Task[T]) Schedule(tx db.Tx, retries int, args T) error {
	key := queueKey(t.name)
	queue, err := db.GetJsonDefault[[]dbTask[T]](tx, key, nil)
	if err != nil {
		return fmt.Errorf("get queue %q: %w", key, err)
	}

	dbt := dbTask[T]{
		Name: t.name,
		TTL:  retries + 1,
		Args: args,
	}
	queue = append(queue, dbt)
	if err := db.SetJson(tx, key, queue); err != nil {
		return fmt.Errorf("set queue %q: %w", key, err)
	}

	select {
	case t.registry.queue.taskEnqueued <- struct{}{}:
	default:
		slog.Info("miss notifying task scheduling", slog.Any("task", dbt))
	}

	slog.Info("scheduled task", slog.Any("task", dbt))
	return nil
}

func RegisterHandler[T any](registry *Registry, name string, handler Handler[T]) (Task[T], error) {
	if _, ok := registry.handlers[name]; ok {
		return Task[T]{}, fmt.Errorf("task handler for %q already registered", name)
	}

	slog.Info("registered task handler", slog.String("name", name))
	registry.handlers[name] = handler
	return Task[T]{
		name:     name,
		registry: registry,
	}, nil
}
