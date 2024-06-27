package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/dgraph-io/badger/v4"
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

func buildDBOpts(path string) badger.Options {
	if path == ":memory:" {
		return badger.DefaultOptions("").WithInMemory(true)
	}

	return badger.DefaultOptions(path).
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
