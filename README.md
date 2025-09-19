# uuid47

[![CI](https://github.com/n2p5/uuid47/actions/workflows/ci.yml/badge.svg)](https://github.com/n2p5/uuid47/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/n2p5/uuid47.svg)](https://pkg.go.dev/github.com/n2p5/uuid47)
[![Go Report Card](https://goreportcard.com/badge/github.com/n2p5/uuid47)](https://goreportcard.com/report/github.com/n2p5/uuid47)
[![License: CC0](https://img.shields.io/badge/License-CC0-lightgrey.svg)](https://creativecommons.org/publicdomain/zero/1.0/)

A Go port of the [UUIDv4/v7 hybrid library](https://github.com/stateless-me/uuidv47) that allows storing time-ordered UUIDv7 in databases while emitting a UUIDv4-looking facade at API boundaries.

## Overview

This library provides a deterministic, invertible mapping between UUIDv7 and UUIDv4-looking UUIDs by XOR-masking only the UUIDv7 timestamp field with a keyed SipHash-2-4 stream tied to the UUID's own random bits.

### Key Features

- **Time-ordered storage**: Store UUIDv7 in your database for efficient indexing and sorting
- **Privacy-preserving API**: Expose UUIDv4-looking identifiers that hide timing patterns
- **Deterministic mapping**: Reversible transformation using a secret key
- **Minimal dependency**: Only depends on `github.com/dchest/siphash` for cryptographic operations
- **Exact C compatibility**: Produces byte-for-byte identical results to the C implementation
- **High performance**: Zero allocations in encode/decode operations
- **Small API surface**: Just 5 public functions focused on the core use case

## Installation

```bash
go get github.com/n2p5/uuid47
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/n2p5/uuid47"
)

func main() {
    // Create a secret key for the transformation
    key := uuid47.Key{K0: 0x0123456789abcdef, K1: 0xfedcba9876543210}
    // Or generate a random key:
    // key, _ := uuid47.NewRandomKey()

    // Parse a UUIDv7 (e.g., from your database)
    v7, _ := uuid47.Parse("018f2d9f-9a2a-7def-8c3f-7b1a2c4d5e6f")

    // Convert to v4 facade for external API
    v4facade := uuid47.Encode(v7, key)
    fmt.Printf("External ID: %s\n", v4facade)
    // Output: External ID: 2463c780-7fca-4def-8c3f-7b1a2c4d5e6f

    // Convert back to v7 for internal use
    original := uuid47.Decode(v4facade, key)
    fmt.Printf("Internal ID: %s\n", original)
    // Output: Internal ID: 018f2d9f-9a2a-7def-8c3f-7b1a2c4d5e6f
}
```

## How It Works

1. **Preserves random bits**: The 74 random bits of UUIDv7 remain unchanged
2. **Masks timestamp**: The 48-bit timestamp is XORed with a SipHash-2-4 stream
3. **Key derivation**: The masking key is derived from the UUID's own random bits
4. **RFC compliance**: Both UUIDs maintain proper version and variant bits

The transformation is:
- **Deterministic**: Same input always produces same output
- **Invertible**: Can recover original UUIDv7 with the secret key
- **Secure**: SipHash-2-4 provides cryptographic security against key recovery

## API Reference

### Types

```go
type UUID [16]byte                  // 128-bit UUID
type Key struct { K0, K1 uint64 }   // 128-bit SipHash key
```

### Core Functions

```go
// Transform between v7 and v4 facade
func Encode(v7 UUID, key Key) UUID
func Decode(v4facade UUID, key Key) UUID

// Generate a random key
func NewRandomKey() (Key, error)

// Parse UUID from string
func Parse(s string) (UUID, error)

// Format UUID as string
func (u UUID) String() string
```

The library implements `fmt.Stringer` for string formatting.

## Compatibility

This Go implementation produces **byte-for-byte identical** output to the [C reference implementation](https://github.com/stateless-me/uuidv47). All test vectors from the C version pass in this implementation.

## Performance

```bash
BenchmarkEncode-10    10000000    9 ns/op     0 B/op    0 allocs/op
BenchmarkDecode-10    10000000    9 ns/op     0 B/op    0 allocs/op
BenchmarkRoundtrip-10  5000000   26 ns/op     0 B/op    0 allocs/op
```

## Security Considerations

- Keep your `Key` secret - it's required to reverse the transformation
- The facade only hides timing information, not the random bits
- Use cryptographically secure random number generation for UUIDv7 creation
- SipHash-2-4 provides 64-bit security against key recovery

## Testing

The library includes comprehensive tests that verify:
- Exact compatibility with C implementation test vectors
- SipHash-2-4 test vectors from the C implementation
- Invertibility property (encode then decode returns original)
- RFC compliance for version and variant bits

Run tests:
```bash
go test -v
```

Run benchmarks:
```bash
go test -bench=.
```

## License

CC0 1.0 Universal - see LICENSE file for details

## Credits

This is a Go port of the excellent [uuidv47](https://github.com/stateless-me/uuidv47) C library by Stateless Limited.