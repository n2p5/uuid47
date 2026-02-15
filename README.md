# uuid47

[![CI](https://github.com/n2p5/uuid47/actions/workflows/ci.yml/badge.svg)](https://github.com/n2p5/uuid47/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/n2p5/uuid47.svg)](https://pkg.go.dev/github.com/n2p5/uuid47)
[![Go Report Card](https://goreportcard.com/badge/github.com/n2p5/uuid47)](https://goreportcard.com/report/github.com/n2p5/uuid47)
[![License: CC0](https://img.shields.io/badge/License-CC0-lightgrey.svg)](https://creativecommons.org/publicdomain/zero/1.0/)

A Go port of the [UUIDv4/v7 hybrid library](https://github.com/stateless-me/uuidv47)
that allows storing time-ordered UUIDv7 in databases while emitting a
UUIDv4-looking facade at API boundaries.

## Overview

This library provides a deterministic, invertible mapping between UUIDv7 and
UUIDv4-looking UUIDs by XOR-masking the UUIDv7 timestamp field with a keyed
SipHash-2-4 stream tied to the UUID's own random bits.

- **Time-ordered storage**: Store UUIDv7 for efficient indexing and sorting
- **Privacy-preserving API**: Expose UUIDv4-looking identifiers that hide
  timing patterns
- **Deterministic and invertible**: Recover the original UUIDv7 with the
  secret key
- **Exact C compatibility**: Byte-for-byte identical output to the
  [C reference implementation](https://github.com/stateless-me/uuidv47)
- **High performance**: Zero allocations in encode, decode, and parse

## Requirements

- Go 1.25 or later

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

## API

```go
type UUID [16]byte                  // 128-bit UUID
type Key struct { K0, K1 uint64 }   // 128-bit SipHash key

func Encode(v7 UUID, key Key) UUID  // v7 -> v4 facade
func Decode(v4 UUID, key Key) UUID  // v4 facade -> v7
func NewRandomKey() (Key, error)    // generate a random key
func Parse(s string) (UUID, error)  // parse UUID string
func (u UUID) String() string       // format UUID string (implements fmt.Stringer)
```

## Performance

```text
BenchmarkEncode-14       130376749       9.036 ns/op      0 B/op    0 allocs/op
BenchmarkDecode-14       135127963       8.894 ns/op      0 B/op    0 allocs/op
BenchmarkRoundtrip-14     41260035      26.52 ns/op       0 B/op    0 allocs/op
BenchmarkParse-14         47770144      25.74 ns/op       0 B/op    0 allocs/op
BenchmarkString-14        51548882      22.36 ns/op      48 B/op    1 allocs/op
```

## Security Considerations

- Keep your `Key` secret — it's required to reverse the transformation
- The facade only hides timing information, not the random bits
- SipHash-2-4 provides 64-bit security against key recovery

## Testing

```bash
make test              # run tests
make bench             # run benchmarks
make verify-c-compat   # verify C compatibility
```

## License

CC0 1.0 Universal — see LICENSE file for details.

## Credits

Go port of the [uuidv47](https://github.com/stateless-me/uuidv47) C library
by Stateless Limited.
