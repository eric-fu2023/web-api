package util

import (
	"math/rand"
	"time"

	"golang.org/x/exp/constraints"
)

type Numeric interface {
	int | int64 | float64
}

func Max[T constraints.Ordered](a T, b T) T {
	if a > b {
		return a
	}
	return b
}

func Sum[T any, V Numeric](objs []T, calcFunc func(T) V) (total V) {
	for _, obj := range objs {
		total += calcFunc(obj)
	}
	return
}

func RandomNumFromRange[T Numeric](start, end T) (res T) {

	rng := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	res = T(rng.Float64()*float64(end-start+1)) + start

	return
}
