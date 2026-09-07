// Demo: generic methods, new in Go 1.27.
// Same List[E] type, but Map is now a method with its
// own type parameter — real method chaining.
package main

import "fmt"

type List[E any] []E

func NewList[E any](items ...E) List[E] {
	return List[E](items)
}

// Map has its own type parameter R, independent of the
// receiver's E — this is the part that required Go 1.27.
func (l List[E]) Map[R any](f func(E) R) List[R] {
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
	result := NewList(0, 2, 4).
		Map(add(2)).
		Map(divideBy(2))
	fmt.Println(result)
}
