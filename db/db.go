package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrKeyNotFound = errors.New("db not found")
)

type DB interface {
	Start(context.Context) error
	Stop()
	Do(context.Context, func(Tx) error) error
}

type Tx interface {
	Get(key []byte) ([]byte, error)
	Set(key []byte, val []byte) error
}

func GetJson[T any](tx Tx, key string) (T, error) {
	data, err := tx.Get([]byte(key))
	if err != nil {
		return *new(T), fmt.Errorf("get key: %w", err)
	}

	result := *new(T)
	err = json.Unmarshal(data, &result)
	if err != nil {
		return *new(T), fmt.Errorf("unmarshall: %w", err)
	}

	return result, nil
}

func GetJsonDefault[T any](tx Tx, key string, val T) (T, error) {
	result, err := GetJson[T](tx, key)
	if errors.Is(err, ErrKeyNotFound) {
		return val, nil
	}

	return result, err
}

func SetJson[T any](tx Tx, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("marshall: %w", err)
	}

	err = tx.Set([]byte(key), data)
	if err != nil {
		return fmt.Errorf("set key: %w", err)
	}

	return nil
}
