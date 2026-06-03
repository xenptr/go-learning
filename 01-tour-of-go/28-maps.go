package main

import "fmt"

// Vertex represents a location using latitude and longitude.
type Vertex struct {
	Lat, Long float64
}

// Declare a map variable.
// Key   -> string (place name)
// Value -> Vertex (coordinates)
var m map[string]Vertex

// map-literals
var ml = map[string]Vertex{
	"Bell Labs": {
		40.68433, -74.39967,
	},
	"Google": {
		37.42202, -122.08408,
	},
}

func main() {
	// make() creates and initializes the map.
	// Without this, m would be nil and we couldn't add entries.
	m = make(map[string]Vertex)

	// Add a key-value pair to the map.
	// Key: "Bell Labs"
	// Value: Vertex containing latitude and longitude.
	m["Bell Labs"] = Vertex{
		40.68433,
		-74.39967,
	}

	// Retrieve and print the value associated with "Bell Labs".
	fmt.Println(m["Bell Labs"])

	fmt.Println(ml)
}
