package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/boar-d-white-foundation/drone/db"
)

const (
	keyAppliedMigrations = "drone:applied_migrations"
)

type appliedMigration struct {
	ID        string    `json:"id"`
	AppliedAt time.Time `json:"applied_at"`
}

type appliedMigrations struct {
	Applied []appliedMigration `json:"applied"`
}

type migration struct {
	id   string
	name string
	fn   func(tx db.Tx) error
}

var migrations = []migration{
	{id: "0001", name: "add_default_greeted_users", fn: addDefaultGreetedUsers},
}

func migrate(ctx context.Context, database db.DB) error {
	slog.Info("applying migrations")
	err := database.Do(ctx, func(tx db.Tx) error {
		applied, err := db.GetJsonDefault(tx, keyAppliedMigrations, appliedMigrations{})
		if err != nil {
			return fmt.Errorf("get applied: %w", err)
		}

		appliedIDs := make(map[string]struct{}, len(applied.Applied))
		for _, m := range applied.Applied {
			appliedIDs[m.ID] = struct{}{}
		}

		for _, mgr := range migrations {
			if _, ok := appliedIDs[mgr.id]; ok {
				slog.Info(
					"migration already applied, skipping",
					slog.String("id", mgr.id), slog.String("name", mgr.name),
				)
				continue
			}

			slog.Info("applying migration", slog.String("id", mgr.id), slog.String("name", mgr.name))
			if err := mgr.fn(tx); err != nil {
				return fmt.Errorf("apply migration %s: %w", mgr.name, err)
			}

			applied.Applied = append(applied.Applied, appliedMigration{
				ID:        mgr.id,
				AppliedAt: time.Now(),
			})
		}

		if err := db.SetJson(tx, keyAppliedMigrations, applied); err != nil {
			return fmt.Errorf("set keyAppliedMigrations: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	slog.Info("migrations applied")
	return nil
}

func addDefaultGreetedUsers(tx db.Tx) error {
	key := "boardwhite:on_join_greeted_users"
	greetedUsers, err := db.GetJsonDefault(tx, key, make(map[int64]struct{}))
	if err != nil {
		return fmt.Errorf("get greetedUsers: %w", err)
	}

	for _, uid := range []int64{142944542} {
		greetedUsers[uid] = struct{}{}
	}
	if err := db.SetJson(tx, key, greetedUsers); err != nil {
		return fmt.Errorf("set keyOnJoinGreetedUsers: %w", err)
	}

	return nil
}
