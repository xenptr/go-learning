package main

import (
	"fmt"
	"io"
	"strings"
)

func main() {
	// Create a Reader from a string
	r := strings.NewReader("Hello, Reader!")

	// Buffer that will receive data
	b := make([]byte, 8)

	for {

		// Read up to len(b) bytes into b
		n, err := r.Read(b)

		fmt.Printf("n = %v err = %v b = %v\n",
			n, err, b)

		// Print only the bytes actually read
		fmt.Printf("b[:n] = %q\n", b[:n])

		// Stop when no more data exists
		if err == io.EOF {
			break
		}
	}
}
