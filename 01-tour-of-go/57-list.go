package main

import "fmt"

// List represents a singly-linked list that holds
// values of any type.
type List[T any] struct {
	next *List[T]
	val  T
}

// Add inserts a new node at the front.
func (l *List[T]) Add(val T) *List[T] {
	return &List[T]{
		val:  val,
		next: l,
	}
}

// Print displays all elements.
func (l *List[T]) Print() {
	for current := l; current != nil; current = current.next {
		fmt.Printf("%v -> ", current.val)
	}
	fmt.Println("nil")
}

func main() {
	// List of ints
	var ints *List[int]

	ints = ints.Add(30)
	ints = ints.Add(20)
	ints = ints.Add(10)

	fmt.Println("Integer List:")
	ints.Print()

	// List of strings
	var strings *List[string]

	strings = strings.Add("baz")
	strings = strings.Add("bar")
	strings = strings.Add("foo")

	fmt.Println("\nString List:")
	strings.Print()
}
