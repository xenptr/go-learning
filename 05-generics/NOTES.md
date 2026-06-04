# 05-generics — Tutorial Notes

Source: https://go.dev/doc/tutorial/generics

---

## What's new here vs the Tour of Go

The Tour (`56-generics.go`, `57-list.go`) covered the basics:

| Tour | Showed |
|---|---|
| `56-generics.go` | Single type parameter `[T comparable]`, used on a slice |
| `57-list.go` | Generic struct `List[T any]` with a method |

This tutorial adds concepts the Tour only touched or skipped entirely:

| New concept | Where |
|---|---|
| Multiple type parameters `[K, V]` | `SumIntsOrFloats` |
| Union constraints `int64 \| float64` | `SumIntsOrFloats` |
| Named constraint interface | `type Number interface { int64 \| float64 }` |
| Type argument inference | calling `SumIntsOrFloats(ints)` without `[string, int64]` |

---

## Project structure

```
05-generics/
├── go.mod      ← module example/generics
├── main.go     ← full annotated program
└── NOTES.md    ← this file
```

---

## Run it

```bash
go run .
```

Expected output:

```
Non-Generic Sums: 46 and 62.97
Generic Sums: 46 and 62.97
Generic Sums, type parameters inferred: 46 and 62.97
Generic Sums with Constraint: 46 and 62.97
```

---

## Key concepts

### Generic function syntax

```
func Name[TypeParams](regularParams) returnType { ... }
```

Everything inside `[...]` is the **type parameter list**. Each entry is:

```
ParameterName Constraint
```

Full example:

```go
//          ┌─ type param ─┐  ┌── type param ───────────┐
func SumIntsOrFloats[K comparable, V int64 | float64](m map[K]V) V {
//                                                    └─ regular param ─┘
```

---

### `comparable` — the built-in constraint

`comparable` is predeclared by Go. It matches any type that supports `==`
and `!=` — things like `int`, `string`, `bool`, pointers, structs with
comparable fields. Slices, maps, and functions are **not** comparable.

You need `comparable` for map keys because Go itself requires them to be
comparable. Without it, `map[K]V` would be rejected by the compiler.

The Tour introduced this in `56-generics.go` with a single type param:

```go
func Index[T comparable](s []T, x T) int { ... }
```

---

### Union constraints `|`

A type parameter constraint can list exact concrete types joined by `|`:

```go
V int64 | float64
```

This means: "at compile time, V must be substituted by exactly int64 or
exactly float64 — nothing else". The compiler enforces this. The `|` here
is NOT a runtime OR — the compiler picks one concrete type per call site.

You can only use operations that are valid for **all** types in the union.
Both `int64` and `float64` support `+=`, so the function body compiles.
If you tried to call a string method on `V`, it would fail because `string`
is not in the union.

---

### Named constraint interfaces

Instead of repeating `int64 | float64` in every function signature, extract
it into a named interface:

```go
type Number interface {
    int64 | float64
}
```

Then use it as a constraint:

```go
func SumNumbers[K comparable, V Number](m map[K]V) V { ... }
```

This is a **constraint interface** — it can only be used as a type parameter
constraint, not as a regular variable type:

```go
var n Number = 42  // compile error — Number is a constraint, not a value type
```

---

### `any` vs union constraints

| Constraint | Meaning | Operations allowed |
|---|---|---|
| `any` | any type at all | only operations valid for ALL types (almost nothing) |
| `comparable` | any type supporting == and != | equality checks only |
| `int64 \| float64` | exactly these two types | arithmetic, comparison |

`any` is the most permissive but also the most restrictive in terms of what
you can actually do with the value. If you need arithmetic, use a union.

---

### Type argument inference

When calling a generic function, you can supply type arguments explicitly:

```go
SumIntsOrFloats[string, int64](ints)
//             └──────────────┘
//             explicit: K=string, V=int64
```

Or let the compiler infer them from the function arguments:

```go
SumIntsOrFloats(ints)
// ints is map[string]int64, so compiler infers K=string, V=int64
```

Inference works when the compiler can determine all type parameters from
the types of the regular arguments. It does **not** work when:
- The generic function has no regular parameters
- The type parameters appear only in the return type

When inference fails, you must supply type arguments explicitly.

---

### Multiple type parameters

You can declare as many type parameters as needed, each with its own
constraint, separated by commas:

```go
func F[A any, B comparable, C int | string](a A, b B, c C) { ... }
```

Each parameter is independent — `A`, `B`, and `C` can be different types at
the call site.

---

### The problem generics solve

Without generics you needed one function per type:

```go
func SumInts(m map[string]int64) int64   { /* logic */ }
func SumFloats(m map[string]float64) float64 { /* same logic */ }
```

This is duplication. Any bug fix or logic change must be applied to every
copy. Generics let you write the logic once and have the compiler generate
the concrete versions for each call site — with full type safety and no
runtime overhead (unlike `interface{}`/`any` approaches, which box values
on the heap and require type assertions).
