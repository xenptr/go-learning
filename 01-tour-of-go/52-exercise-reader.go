package main

import "golang.org/x/tour/reader"

type MyReader struct{}

func (r MyReader) Read(b []byte) (int, error) {
	// Fill the provided buffer with 'A'
	for i := range b {
		b[i] = 'A'
	}

	// All bytes were written successfully.
	// Since the stream is infinite, we never return io.EOF.
	return len(b), nil
}

func main() {
	reader.Validate(MyReader{})
}
