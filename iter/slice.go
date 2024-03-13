package iter

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func PickRandom[T any](xs []T) (T, error) {
	idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(xs))))
	if err != nil {
		return *new(T), fmt.Errorf("generate random: %w", err)
	}
	return xs[idx.Int64()], nil
}
