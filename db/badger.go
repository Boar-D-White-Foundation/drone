package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/boar-d-white-foundation/drone/config"
	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
)

type BadgerDB struct {
	sync.Mutex
	badgerOpts badger.Options
	bdb        *badger.DB
}

type BadgerTx struct {
	btx *badger.Txn
}

var _ DB = (*BadgerDB)(nil)
var _ Tx = (*BadgerTx)(nil)

func NewBadgerDB(path string) *BadgerDB {
	badgerOpts := buildDBOpts(path)
	return &BadgerDB{
		badgerOpts: badgerOpts,
	}
}

func NewBadgerDBFromConfig(cfg config.Config) DB {
	return NewBadgerDB(cfg.BadgerPath)
}

type slogLogger struct{}

func (l slogLogger) Errorf(msg string, args ...interface{}) {
	slog.Error(fmt.Sprintf(msg, args...), slog.String("logger", "badger"))
}

func (l slogLogger) Warningf(msg string, args ...interface{}) {
	slog.Warn(fmt.Sprintf(msg, args...), slog.String("logger", "badger"))
}

func (l slogLogger) Infof(msg string, args ...interface{}) {
	slog.Info(fmt.Sprintf(msg, args...), slog.String("logger", "badger"))
}

func (l slogLogger) Debugf(msg string, args ...interface{}) {
	slog.Debug(fmt.Sprintf(msg, args...), slog.String("logger", "badger"))
}

func buildDBOpts(path string) badger.Options {
	if path == ":memory:" {
		return badger.DefaultOptions("").WithInMemory(true)
	}

	return badger.DefaultOptions(path).
		WithLogger(slogLogger{}).
		WithSyncWrites(true).
		WithVerifyValueChecksum(true).
		WithChecksumVerificationMode(options.OnTableAndBlockRead).
		// https://github.com/dgraph-io/badger/issues/1297#issuecomment-612941482
		WithValueLogFileSize(1024 * 1024 * 16).
		WithNumVersionsToKeep(1).
		WithCompactL0OnClose(true).
		WithNumLevelZeroTables(1).
		WithNumLevelZeroTablesStall(2)
}

func (b *BadgerDB) Start(ctx context.Context) error {
	bdb, err := badger.Open(b.badgerOpts)
	if err != nil {
		return fmt.Errorf("open badger: %w", err)
	}

	b.bdb = bdb
	return nil
}

func (b *BadgerDB) Stop() {
	if err := b.bdb.Close(); err != nil {
		slog.Error("failed to close badger", err)
	}
}

func (b *BadgerDB) Do(ctx context.Context, f func(Tx) error) error {
	b.Lock()
	defer b.Unlock()

	err := b.bdb.Update(func(btx *badger.Txn) error {
		tx := BadgerTx{btx: btx}
		return f(&tx)
	})
	if err != nil {
		return fmt.Errorf("badger tx: %w", err)
	}

	return nil
}

func (b *BadgerDB) Dump(ctx context.Context) ([]KV, error) {
	// badger already has a file level lock
	b.Lock()
	defer b.Unlock()

	var dump []KV
	err := b.bdb.View(func(btx *badger.Txn) error {
		it := btx.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)
			val, err := item.ValueCopy(nil)
			if err != nil {
				return fmt.Errorf("get value: %w", err)
			}

			dump = append(dump, KV{
				Key: key,
				Val: val,
			})
			slog.Info("dump key", slog.String("key", string(key)))
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("badger tx: %w", err)
	}

	return dump, nil
}

func (tx *BadgerTx) Get(key []byte) ([]byte, error) {
	item, err := tx.btx.Get(key)
	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, ErrKeyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get badger key %q: %w", key, err)
	}

	data, err := item.ValueCopy(nil)
	if err != nil {
		return nil, fmt.Errorf("get value %w", err)
	}

	return data, nil
}

func (tx *BadgerTx) Set(key []byte, val []byte) error {
	err := tx.btx.Set(key, val)
	if err != nil {
		return fmt.Errorf("set badger key %q: %w", key, err)
	}

	return nil
}
