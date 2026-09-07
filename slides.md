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

# The FAQ Said It Best

> "We do not anticipate that Go will ever add generic methods."

— Go FAQ, for most of the last decade

---

# Then, Go 1.27 Shipped This

```go
func (List[E]) Map[R any](f func(E) R) List[R] {
    // ...
}

list.Map(add(2)).Map(divideBy(2))
```

That's the talk.

---

# Agenda

- Where generics started (Go 1.18)
- The quiet years (1.19 → 1.26)
- What changed in Go 1.27
- Generic methods, in depth
- Why generic *interface* methods still can't —
  the separate compilation problem, from the Go blog
- Generalized function type inference
- Performance: why this is "free"
- Recap & discussion

---

# 2022 — Go 1.18: Generics Launch

- Type parameters on **functions and types**
- `any` (alias for `interface{}`), `comparable`
- Constraints: unions (`int | float64`), approximation (`~string`)
- "Core type" concept governs what ops are legal

```go
func Foo[T any](x T) T { return x }
```

Methods were **explicitly excluded** from type parameters.
Why, exactly? We'll come back to that once you've
seen what generic methods actually look like.

---

# The Quiet Years (2023 → 2026)

Generics didn't stand still — just nothing anyone
was actually asking for:

- **1.21** — stdlib catches up: `slices`, `maps`, `cmp`
- **1.23** — iterators: range-over-func, `iter.Seq`
- **1.24** — generic type aliases finally work
- **1.25** — "core types" quietly removed from the spec
  (a leak in the 1.18 abstraction, fixed)
- **1.26** — self-referential type parameters

Useful. Incremental. Not the big ask.

---

# What Everyone Was Actually Asking For

Proposal **#49085** — "allow type parameters in methods"

Filed October 2021.
900+ 👍 reactions.

Rejected. For four years.

---

# Go 1.27 — August 19, 2026

Two language-level generics changes. That's it.

1. **Generic methods**
2. **Generalized function type inference**

(A third, unrelated change also shipped — struct literal
field selectors. Not our topic for this talk.)

No new constraint syntax. No generic type alias changes.
Both are small in surface area, large in consequence.

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

# Stdlib Example: math/rand/v2

**Before** — no method existed at all. Methods couldn't be
generic, so `N` only worked on the package's global source:

```go
func N[Int intType](n Int) Int
rand.N[int](10) // always the global Rand
```

Your own seeded `*Rand`? No generic `N` for it.

**After** — a real generic method, on any `*Rand`:

```go
func (r *Rand) N[Int intType](n Int) Int
myRand.N[int](10) // your seed, your instance
```

---

# So Why Was This Excluded Until Now?

Time to answer the question from the 1.18 slide.

Two packages, one interface:

```text
 package main              package p
 ───────────────           ─────────────────────
 type T struct{}           type I interface { M() }
 func (T) M() {...}        func F(i I) { i.M() }

 main() { p.F(T{}) } ────► F(i I)
                              │
                           i.M()  ── routed to T.M
                                     at runtime
```

Compiled **separately**. Package `p` never sees `T`.
It only knows "something implementing `I`" arrives.

*Source: go.dev/blog/generic-methods,
"The trouble with generic interface methods"*

---

# Non-Generic Methods: Easy Mode

`p` can't know how it will use `T{}` — so Go doesn't
try to guess. It just compiles **every** method of `T`
the moment `T` is declared.

```text
   declare T  ──►  compile T.M (and all its methods)
                        │
              guaranteed to exist at runtime,
              no matter what p.F does with it
```

One method. One compiled body. Done.

---

# Now Make `M` Generic

```go
type T struct{}
func (T) M[P any]() { /* ... */ }

type I interface { M[P any]() }
func F(i I) { i.M[int]() }   // ← but could be M[string]? M[Foo]?
```

`p` decides the type argument. `T`'s package doesn't
know it. To stay safe, the compiler would need to
pre-generate **every possible instantiation**:

```text
T.M[int]   T.M[string]   T.M[Foo]   T.M[...]   ...
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

# Proposal History

- `#49085` (2021) — rejected, "generic interface
  methods look impractical, so why bother at all"
- `#77273` (Jan 2026) — Robert Griesemer reopens it,
  separates the interface question from the
  organizational-code-reuse question
- Accepted. Shipped in Go 1.27.

---

# Function Type Inference: The Old Gaps

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

# Performance: Why This Is "Free"

A generic method call on a concrete type is resolved
**statically, at compile time** — the compiler knows
the receiver's type and the type argument, so it
generates the same kind of call it always did.

No interface, no itable, no runtime type resolution
involved.

Theoretically: no runtime cost beyond what generics
already cost before 1.27.

---

# Recap

- Go 1.27 shipped **two** generics changes
- Generic methods: real method chaining, four years
  in the making
- Interfaces still can't have generic methods —
  that limitation is unchanged, and deliberate
- Function type inference: one consistent rule instead
  of a patchwork of special cases

---

# Discussion

Does this move Go toward Rust/Java-level complexity —

or is it a scoped, overdue fix to a four-year-old gap?

---

# Thanks, GoWroc

Questions?
