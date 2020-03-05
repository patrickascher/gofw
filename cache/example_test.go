package cache_test

import (
	"github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/cache/memory"
)

// These examples demonstrate the basics of printing using a format string. Printf,
// Sprintf, and Fprintf all take a format string that specifies how to format the
// subsequent arguments. For example, %d (we call that a 'verb') says to print the
// corresponding argument, which must be an integer (or something containing an
// integer, such as a slice of ints) in decimal. The verb %v ('v' for 'value')
// always formats the argument in its default form, just how Print or Println would
// show it. The special verb %T ('T' for 'Type') prints the type of the argument
// rather than its value. The examples are not exhaustive; see the package comment
// for all the details.
func Example_register() {
	// A basic set of examples showing that %v is the default format, in this
	// case decimal for integers, which can be explicitly requested with %d;
	// the output is just what Println generates.
	err := cache.Register("memory", memory.New)
	if err != nil {
		// ...
	}
}

func ExampleRegister() {
	// A basic set of examples showing that %v is the default format, in this
	// case decimal for integers, which can be explicitly requested with %d;
	// the output is just what Println generates.
	err := cache.Register(cache.MEMORY, memory.New)
	if err != nil {
		// ...
	}
}
