// Demo: generic methods, new in Go 1.27.
// Result[T] gets a Map method with its own type parameter R,
// independent of the receiver's T — real chaining for
// validation/parsing pipelines.
package main

import (
	"fmt"
	"strconv"
)

type Result[T any] struct {
	val T
	err error
}

func Parse(s string) Result[string] {
	return Result[string]{val: s}
}

// Map has its own type parameter R, independent of the
// receiver's T — this is the part that required Go 1.27.
func (r Result[T]) Map[R any](f func(T) (R, error)) Result[R] {
	if r.err != nil {
		return Result[R]{err: r.err}
	}
	out, err := f(r.val)
	return Result[R]{val: out, err: err}
}

func double(n int) (int, error) {
	return n * 2, nil
}

func main() {
	parsed := Parse("21").
		Map(strconv.Atoi).
		Map(double)

	if parsed.err != nil {
		fmt.Println("error:", parsed.err)
		return
	}
	fmt.Println(parsed.val)
}
