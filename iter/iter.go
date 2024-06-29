package iter

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

func PickRandom[T any](xs []T) (T, error) {
	idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(xs))))
	if err != nil {
		return *new(T), fmt.Errorf("generate random: %w", err)
	}
	return xs[idx.Int64()], nil
}

func FilterMut[T any](xs []T, f func(T) bool) []T {
	insertIdx := 0
	for _, x := range xs {
		if f(x) {
			xs[insertIdx] = x
			insertIdx++
		}
	}
	return xs[:insertIdx]
}

func JoinNonEmpty(sep string, xs ...string) string {
	nonEmpty := make([]string, 0, len(xs))
	for _, s := range xs {
		if len(s) > 0 {
			nonEmpty = append(nonEmpty, s)
		}
	}
	return strings.Join(nonEmpty, sep)
}
