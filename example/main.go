// Example program demonstrating UUIDv4/v7 hybrid functionality.
// This is a port of the C demo.c program.
package main

import (
	"fmt"

	"github.com/n2p5/uuid47"
)

func main() {
	// Example: parse a v7 from DB, emit facade, then decode back.
	key := uuid47.Key{K0: 0x0123456789abcdef, K1: 0xfedcba9876543210}

	// Example v7 string (any valid v7 will do):
	s := "018f2d9f-9a2a-7def-8c3f-7b1a2c4d5e6f"
	idV7, err := uuid47.Parse(s)
	if err != nil {
		fmt.Printf("Error parsing UUID: %v\n", err)
		return
	}

	// Encode to v4 facade
	facade := uuid47.Encode(idV7, key)

	// Decode back to v7
	back := uuid47.Decode(facade, key)

	fmt.Printf("v7 in : %s\n", idV7)
	fmt.Printf("v4 out: %s\n", facade)
	fmt.Printf("back  : %s\n", back)
}
