package iterx

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

func Uniq[T comparable](xs []T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0, len(xs))
	for _, x := range xs {
		if _, ok := seen[x]; !ok {
			seen[x] = struct{}{}
			result = append(result, x)
		}
	}
	return result
}

func JoinNonEmpty(sep string, xs ...string) string {
	nonEmpty := FilterMut(xs, func(s string) bool { return len(s) > 0 })
	return strings.Join(nonEmpty, sep)
}
