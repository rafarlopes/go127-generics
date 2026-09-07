// Demo: generalized function type inference, new in Go 1.27.
// Before 1.27, assigning a generic function into a struct
// literal required an explicit type argument: S{f: identity[int]}.
package main

import "fmt"

func identity[T any](v T) T {
	return v
}

type S struct {
	f func(int) int
}

func main() {
	// Go 1.27: inferred from the field's type, no identity[int] needed.
	s := S{f: identity}
	fmt.Println(s.f(42))

	// Same rule now also applies to channel sends and conversions.
	ch := make(chan func(int) int, 1)
	ch <- identity
	fmt.Println((<-ch)(7))
}
