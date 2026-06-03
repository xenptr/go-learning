package main

import (
	"io"
	"os"
	"strings"
)

// rot13Reader wraps another Reader.
// It reads data from that Reader and applies ROT13
// before returning the data to the caller.
type rot13Reader struct {
	r io.Reader
}

// Read satisfies the io.Reader interface.
//
// 1. Read data from the wrapped Reader.
// 2. Transform the bytes using ROT13.
// 3. Return the modified data.
func (rot *rot13Reader) Read(b []byte) (int, error) {
	// Fill buffer with data from the underlying Reader.
	n, err := rot.r.Read(b)

	// Only transform the bytes that were actually read.
	if n > 0 {
		for i := 0; i < n; i++ {
			b[i] = rot13(b[i])
		}
	}

	// Return number of bytes read and any error from
	// the underlying Reader (including io.EOF).
	return n, err
}

// Apply ROT13 to a single byte.
//
// A-Z -> N-Z A-M
// a-z -> n-z a-m
//
// Non-alphabetic characters are unchanged.
func rot13(b byte) byte {
	switch {
	case 'A' <= b && b <= 'Z':
		return 'A' + (b-'A'+13)%26

	case 'a' <= b && b <= 'z':
		return 'a' + (b-'a'+13)%26

	default:
		return b
	}
}

func main() {
	// rot13Reader expects another io.Reader as its data source.
	//
	// strings.NewReader turns the string into a Reader so that
	// rot13Reader has something to read from and transform.
	s := strings.NewReader("Lbh penpxrq gur pbqr!")

	// Wrap the Reader with our ROT13 Reader.
	r := rot13Reader{s}

	// Read from r and write to standard output.
	//
	// io.Copy repeatedly calls:
	//     r.Read(...)
	//
	// until io.EOF is returned.
	io.Copy(os.Stdout, &r)
}
