package main

import "github.com/dgraph-io/badger/v4"

func NewBadger(path string) (*badger.DB, error) {
	dbOpts := badger.DefaultOptions(path).
		// https://github.com/dgraph-io/badger/issues/1297#issuecomment-612941482
		WithValueLogFileSize(1024 * 1024 * 16).
		WithNumVersionsToKeep(1).
		WithCompactL0OnClose(true).
		WithNumLevelZeroTables(1).
		WithNumLevelZeroTablesStall(2)

	return badger.Open(dbOpts)
}
