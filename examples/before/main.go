// Demo: generics before Go 1.27 — no generic methods.
// Package-level generic functions are the only option,
// so composing transformations reads inside-out.
package main

import "fmt"

type List[E any] []E

func NewList[E any](items ...E) List[E] {
	return List[E](items)
}

func MapList[E, R any](l List[E], f func(E) R) List[R] {
	out := make(List[R], len(l))
	for i, v := range l {
		out[i] = f(v)
	}
	return out
}

func add(n int) func(int) int {
	return func(x int) int { return x + n }
}

func divideBy(n int) func(int) int {
	return func(x int) int { return x / n }
}

func main() {
	result := MapList(
		MapList(NewList(0, 2, 4), add(2)),
		divideBy(2),
	)
	fmt.Println(result)
}
