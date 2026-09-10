---
theme: ./styles/dracula.json
author: Rafael Lopes
date: MMMM dd, YYYY
paging: "%d / %d"
---

# Go 1.27
## The Generics Update

GoWroc #64

---

# Agenda

- Where generics started (Go 1.18)
- Generics over the years (1.19 → 1.26)
- What changed in Go 1.27
- Generic methods
- Why we cannot still use generic *interface* methods
- Why this is "free"
- Generalized function type inference
- Discussion & Questions

---

# 2022 — Go 1.18: Generics Launch

- Type parameters on **functions and types**
- `any` (alias for `interface{}`), `comparable`
- Type Constraints: `int | float64`, `~string`

```go
func Foo[T any](x T) T { return x }
```

Methods were **explicitly excluded** from type parameters.
Why, exactly? We'll come back to that once you've
seen what generic methods actually look like.

---

# Generics Over The Years (2023 → 2026)

- **1.21** — stdlib catches up: `slices`, `maps`, `cmp`
```go
slices.Contains(s, x)
```

- **1.23** — iterators: range-over-func, `iter.Seq`
```go
for v := range slices.Values(s)
```

- **1.24** — generic type aliases finally work
```go
type Set[T comparable] = map[T]struct{}
```

- **1.26** — self-referential type parameters
```go
type Ordered[T Ordered[T]] interface { Less(T) bool }
```

---

# Go 1.27 — August 19, 2026

Two generics changes:

1. **Generic methods**
2. **Generalized function type inference**

---

# Generic Methods: Before

`examples/before/main.go` — reads inside-out

`go run ./examples/before`

```go
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

func main() {
    result := MapList(
        MapList(NewList(0, 2, 4), add(2)),
        divideBy(2),
    )
    fmt.Println(result) // [1 2 3]
}
```

---

# Generic Methods: After

`examples/after/main.go` — real method chaining

```go
type List[E any] []E

func NewList[E any](items ...E) List[E] {
    return List[E](items)
}

// Map has its own type parameter R, independent of
// the receiver's E — this required Go 1.27.
func (l List[E]) Map[R any](f func(E) R) List[R] {
    out := make(List[R], len(l))
    for i, v := range l {
        out[i] = f(v)
    }
    return out
}

func main() {
    result := NewList(0, 2, 4).
        Map(add(2)).
        Map(divideBy(2))
    fmt.Println(result) // [1 2 3]
}
```

`go run ./examples/after`

---

# So Why Was This Excluded Until Now?

Time to answer the question from the 1.18 slide.

Two packages, one interface:

```text
 package main                    package billing
 ───────────────                 ─────────────────────────────
 type Wallet struct{}            type Balancer interface { Balance() int }
 func (Wallet) Balance() int     func Report(b Balancer) { b.Balance() }
 {...}

 main() { billing.Report(Wallet{}) } ────► Report(b Balancer)
                                                │
                                             b.Balance()  ── routed to
                                                             Wallet.Balance
                                                             at runtime
```

Compiled **separately**. Package `billing` never sees `Wallet`.
It only knows "something implementing `Balancer`" arrives.

*Source: go.dev/blog/generic-methods,
"The trouble with generic interface methods"*

---

# Non-Generic Methods: Easy Mode

`billing` can't know how it will use `Wallet{}` — so Go
doesn't try to guess. It just compiles **every** method
of `Wallet` the moment `Wallet` is declared.

```text
   declare Wallet  ──►  compile Wallet.Balance (and all its methods)
                             │
              guaranteed to exist at runtime,
              no matter what billing.Report does with it
```

One method. One compiled body. Done.

---

# Now Add A Generic Method

```go
type Store struct{ /* ... */ }
func (Store) Get[T any](key string) T { /* ... */ }

type Getter interface { Get[T any](key string) T }
func Report(g Getter) { g.Get[User]("u:42") }   // ← but could be Get[Order]? Get[Invoice]?
```

`billing` decides the type argument. `Store`'s package
doesn't know it. To stay safe, the compiler would need
to pre-generate **every possible instantiation**:

```text
Store.Get[User]   Store.Get[Order]   Store.Get[Invoice]   ...
```

Unbounded. Impractical. That's the wall.

---

# Why Our `Map` Example Was Fine

```go
list.Map[string](f)
//   ▲              ▲
//   │              type argument — known
//   │              right here, at the call site
```

No interface box. No unseen caller in another package
deciding the type argument later.

The compiler only ever compiles the instantiations
that are **actually, directly called** — same as any
other generic function since Go 1.18.

That's the trade: generic concrete methods, yes.
Generic interface methods — still no.

---

# The Framing Shift

Go 1.18: methods exist mainly as an
**interface-implementation mechanism**.

> "If one views methods also as an organizational tool,
> then Go 1.18's reasoning appears overly restrictive."

— Mark Freeman, go.dev/blog/generic-methods

---

# Why This Is "Free"

A generic method call on a concrete type is resolved
**statically, at compile time** — the compiler knows
the receiver's type and the type argument, so it
generates the same kind of call it always did.

Theoretically: no runtime cost beyond what generics
already cost before 1.27.

---

# Switching Gears

That's generic methods — one of Go 1.27's two
language changes.

The second is smaller, quieter, and has nothing
to do with methods: **generalized function type
inference**.

---

# Function Type Inference: The Old Gaps

```go
func g[T any](v T) T { return v }
type IntFormatter func(int) string
```

Worked since Go 1.21 — but only in some contexts:

| Context              | 1.18–1.26        |
|-----------------------|-------------------|
| `s.f = g`              | ✅ works           |
| `s = S{f: g}`          | ❌ needed `g[int]` |
| `c <- g`               | ❌ needed `g[int]` |
| `IntFormatter(g)`      | ❌ needed `g[int]` |

---

# Function Type Inference: Go 1.27

`examples/inference/main.go` — all inferred, no `[int]` anywhere

```go
func identity[T any](v T) T {
    return v
}

type S struct {
    f func(int) int
}

func main() {
    // used to require: S{f: identity[int]}
    s := S{f: identity}
    fmt.Println(s.f(42)) // 42

    // same rule now also applies to channel sends
    ch := make(chan func(int) int, 1)
    ch <- identity
    fmt.Println((<-ch)(7)) // 7
}
```

`go run ./examples/inference`

One rule, applied everywhere a generic function
meets a concrete function type.

---

# Discussion

Does this move Go toward Rust/Java-level complexity or is it a scoped overdue fix?

---

# Thanks, GoWroc

Questions?


Slides:

```
█ ▄▄▄▄▄ █▀█ █▄▀▀▀ █ ▄▄█ ▄▄▄▄▄ █
█ █   █ █▀▀▀█ ▀ ▄▄▀█▄▄█ █   █ █
█ █▄▄▄█ █▀ █▀▀█▀▀▀▀▀▄ █ █▄▄▄█ █
█▄▄▄▄▄▄▄█▄▀ ▀▄█ █ █ █▄█▄▄▄▄▄▄▄█
█▄▄▄ ▄▀▄  ▄▀▄▀ ▄▄▄ ▀▀▀▄▀ ▀▄█▄▀█
█▀▄█▄█ ▄█ █▄█▀▄▄▄██▀▀████▄▀█▀██
█   ▀  ▄█▄▄▄█▄▄▄▄▄ ▀▀ ▀▀▀▀▄▄█▀█
█▀▄  ▀█▄  ▄  ▄█  ▄▄█▀  ▀▀ ▄▄▀██
█▀▄ ▀▄ ▄█▀ █▄▀▄ ▄▄▄█ █ ▀ ▀▄ █▀█
█ █▀▄█ ▄▄█▄▄█▀▄█▄ █ █ ▀ ▄▄█▄▀██
█▄██▄▄▄▄█ ▄▄█▄ ▄█▀▄▀▄ ▄▄▄ ▀   █
█ ▄▄▄▄▄ █▄   ▄█▀▄ ██  █▄█ ▄▄▀██
█ █   █ █ ██▄ ▄▄▄▀▄█▀ ▄▄▄▄▀ ▀ █
█ █▄▄▄█ █ █▀▀▀▄▀▄ ▄▄▄  ▄ ▄ ▄ ██
█▄▄▄▄▄▄▄█▄█▄██▄▄██▄▄▄██▄▄▄█▄███
```

github.com/rafarlopes/go127-generics
