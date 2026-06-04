// Package main demonstrates Go generics beyond what the Tour of Go covers.
//
// Source: https://go.dev/doc/tutorial/generics
//
// The Tour of Go (56-generics.go, 57-list.go) showed:
//   - A single type parameter with the built-in `comparable` constraint
//   - A generic struct (List[T any])
//
// This tutorial adds:
//   - Multiple type parameters in one function: [K comparable, V int64|float64]
//   - Union type constraints: V int64 | float64
//   - Named constraint interfaces: type Number interface { int64 | float64 }
//   - Type argument inference: calling SumIntsOrFloats(ints) without [string, int64]
//
// Run:
//
//	go run .
//
// Expected output:
//
//	Non-Generic Sums: 46 and 62.97
//	Generic Sums: 46 and 62.97
//	Generic Sums, type parameters inferred: 46 and 62.97
//	Generic Sums with Constraint: 46 and 62.97
package main

import "fmt"

// ============================================================
// Named type constraint
// ============================================================
//
// A type constraint is declared as an interface.
// When the interface body contains concrete types joined by |, it means
// "a type parameter using this constraint may be any of these types".
//
// This is different from a regular interface (which lists methods):
//   - Regular interface  → any type that implements those methods
//   - Union interface    → only the listed concrete types exactly
//
// Declaring it here (outside the function) lets us reuse it in multiple
// places instead of repeating `int64 | float64` everywhere.
// It also makes the intent self-documenting: "Number means int64 or float64".
type Number interface {
	int64 | float64
}

func main() {
	// map literals: map[KeyType]ValueType{ key: value, ... }
	ints := map[string]int64{
		"first":  34,
		"second": 12,
	}
	floats := map[string]float64{
		"first":  35.98,
		"second": 26.99,
	}

	// ── Step 1: non-generic, two separate functions ──────────────────────
	// Before generics, you needed one function per concrete type.
	// Both SumInts and SumFloats have identical logic – only the type differs.
	fmt.Printf("Non-Generic Sums: %v and %v\n",
		SumInts(ints),
		SumFloats(floats))

	// ── Step 2: generic, explicit type arguments ─────────────────────────
	// SumIntsOrFloats[K, V] works for both maps with a single implementation.
	// Here we spell out the type arguments in square brackets: [string, int64].
	// K=string (the map key type), V=int64 (the map value type).
	fmt.Printf("Generic Sums: %v and %v\n",
		SumIntsOrFloats[string, int64](ints),
		SumIntsOrFloats[string, float64](floats))

	// ── Step 3: generic, inferred type arguments ─────────────────────────
	// The compiler can deduce K and V from the type of the argument we pass.
	// Because `ints` is map[string]int64, the compiler infers K=string, V=int64.
	// Inference doesn't always work (e.g. when a generic function has no
	// arguments), but when it does it makes call sites cleaner.
	fmt.Printf("Generic Sums, type parameters inferred: %v and %v\n",
		SumIntsOrFloats(ints),
		SumIntsOrFloats(floats))

	// ── Step 4: generic, named constraint ────────────────────────────────
	// SumNumbers uses the Number interface instead of the inline union.
	// Same behaviour, but the constraint is reusable across functions.
	fmt.Printf("Generic Sums with Constraint: %v and %v\n",
		SumNumbers(ints),
		SumNumbers(floats))
}

// ============================================================
// Non-generic functions (before generics existed)
// ============================================================

// SumInts adds together the values of map m and returns the total.
// Only works for map[string]int64 — can't be called with float64 maps.
func SumInts(m map[string]int64) int64 {
	var s int64
	for _, v := range m {
		s += v
	}
	return s
}

// SumFloats adds together the values of map m and returns the total.
// Identical logic to SumInts — only the type is different.
// This duplication is exactly what generics eliminate.
func SumFloats(m map[string]float64) float64 {
	var s float64
	for _, v := range m {
		s += v
	}
	return s
}

// ============================================================
// Generic function — inline union constraint
// ============================================================
//
// Syntax breakdown:  SumIntsOrFloats[K comparable, V int64 | float64]
//
//   [K comparable, V int64 | float64]  — type parameter list (in square brackets)
//   K comparable                       — type param K must support == and !=
//   V int64 | float64                  — type param V must be int64 OR float64
//   (m map[K]V)                        — ordinary parameter using those type params
//   V                                  — return type is also V
//
// K must be comparable because Go requires map keys to be comparable.
// If we used `any` for K, the compiler would reject `map[K]V` as invalid.
//
// The | operator in constraints is a union — it means "this OR that type".
// It does NOT mean the value is an interface holding one of these types;
// at compile time the compiler substitutes one concrete type for V.

// SumIntsOrFloats sums the values of map m.
// It supports both int64 and float64 as map value types.
func SumIntsOrFloats[K comparable, V int64 | float64](m map[K]V) V {
	var s V // zero value of V: 0 for int64, 0.0 for float64
	for _, v := range m {
		s += v
	}
	return s
}

// ============================================================
// Generic function — named constraint
// ============================================================
//
// V Number is equivalent to V int64 | float64 — just using the named
// interface declared above. Prefer named constraints when:
//   - The same union is used in more than one function
//   - The union is long or non-obvious and a name clarifies intent
//
// The function body is identical to SumIntsOrFloats — the only change
// is the type constraint for V.

// SumNumbers sums the values of map m.
// It supports both integers and floats as map values via the Number constraint.
func SumNumbers[K comparable, V Number](m map[K]V) V {
	var s V
	for _, v := range m {
		s += v
	}
	return s
}
