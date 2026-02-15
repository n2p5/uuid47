// Package uuid47 provides UUIDv4/v7 hybrid functionality that allows storing
// time-ordered UUIDv7 in databases while emitting a UUIDv4-looking facade at
// API boundaries. It uses XOR-masking of the UUIDv7 timestamp field with a
// keyed SipHash-2-4 stream tied to the UUID's own random bits.
package uuid47

import (
	"crypto/rand"
	"encoding/binary"
	"errors"

	"github.com/dchest/siphash"
)

// UUID represents a 128-bit UUID.
type UUID [16]byte

// Key represents a 128-bit SipHash key.
type Key struct {
	K0, K1 uint64
}

// ErrInvalidUUID is returned when parsing an invalid UUID string.
var ErrInvalidUUID = errors.New("invalid UUID format")

// Parse parses a UUID string in the format xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func Parse(s string) (UUID, error) {
	var u UUID
	if len(s) != 36 {
		return u, ErrInvalidUUID
	}

	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return u, ErrInvalidUUID
	}

	// Decode hex nibbles directly, skipping validated hyphen positions.
	j := 0
	for i := 0; i < 36; {
		if s[i] == '-' {
			i++
			continue
		}
		hi, ok := hexNibble(s[i])
		if !ok {
			return u, ErrInvalidUUID
		}
		lo, ok := hexNibble(s[i+1])
		if !ok {
			return u, ErrInvalidUUID
		}
		u[j] = (hi << 4) | lo
		j++
		i += 2
	}
	return u, nil
}

// String returns the canonical string representation of a UUID.
func (u UUID) String() string {
	const hexdigits = "0123456789abcdef"
	var buf [36]byte

	// Format as 8-4-4-4-12
	j := 0
	for i := range 16 {
		if i == 4 || i == 6 || i == 8 || i == 10 {
			buf[j] = '-'
			j++
		}
		buf[j] = hexdigits[(u[i]>>4)&0xF]
		buf[j+1] = hexdigits[u[i]&0xF]
		j += 2
	}

	return string(buf[:])
}

// NewRandomKey generates a cryptographically secure random key.
func NewRandomKey() (Key, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return Key{}, err
	}
	return Key{
		K0: binary.LittleEndian.Uint64(buf[0:8]),
		K1: binary.LittleEndian.Uint64(buf[8:16]),
	}, nil
}

// Encode converts a UUIDv7 to a UUIDv4-looking facade.
func Encode(uuid UUID, key Key) UUID {
	// 1) mask = SipHash24(key, v7.random74bits) -> take low 48 bits
	sipMsg := buildSipInputFromV7(uuid)
	mask48 := siphash.Hash(key.K0, key.K1, sipMsg[:]) & 0x0000FFFFFFFFFFFF

	// 2) encTS = ts ^ mask
	ts48 := rd48be(uuid[:6])
	encTS := ts48 ^ mask48

	// 3) build v4 facade: write encTS, set ver=4, keep rand bytes identical, set variant
	out := uuid
	wr48be(out[:6], encTS)
	setVersion(&out, 4)     // facade v4
	setVariantRFC4122(&out) // ensure RFC variant bits
	return out
}

// Decode reverses the facade, recovering the original UUIDv7.
func Decode(uuid UUID, key Key) UUID {
	// 1) rebuild same Sip input from facade (identical bytes)
	sipMsg := buildSipInputFromV7(uuid)
	mask48 := siphash.Hash(key.K0, key.K1, sipMsg[:]) & 0x0000FFFFFFFFFFFF

	// 2) ts = encTS ^ mask
	encTS := rd48be(uuid[:6])
	ts48 := encTS ^ mask48

	// 3) restore v7: write ts, set ver=7, set variant
	out := uuid
	wr48be(out[:6], ts48)
	setVersion(&out, 7)
	setVariantRFC4122(&out)
	return out
}

// Internal helper functions

// hexNibble converts an ASCII hex character to its 4-bit value.
func hexNibble(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	default:
		return 0, false
	}
}

// rd48be reads a 48-bit big-endian value from 6 bytes.
func rd48be(src []byte) uint64 {
	return (uint64(src[0]) << 40) |
		(uint64(src[1]) << 32) |
		(uint64(src[2]) << 24) |
		(uint64(src[3]) << 16) |
		(uint64(src[4]) << 8) |
		uint64(src[5])
}

// wr48be writes a 48-bit value as big-endian to 6 bytes.
func wr48be(dst []byte, v48 uint64) {
	dst[0] = byte(v48 >> 40)
	dst[1] = byte(v48 >> 32)
	dst[2] = byte(v48 >> 24)
	dst[3] = byte(v48 >> 16)
	dst[4] = byte(v48 >> 8)
	dst[5] = byte(v48)
}

// buildSipInputFromV7 extracts exactly the random bits of v7 (rand_a 12b + rand_b 62b)
// as easy full bytes: [low-nibble of b6][b7][b8&0x3F][b9..b15]
// This matches the C implementation's build_sip_input_from_v7 function.
func buildSipInputFromV7(u UUID) [10]byte {
	var msg [10]byte
	msg[0] = u[6] & 0x0F // low nibble of byte 6
	msg[1] = u[7]        // byte 7
	msg[2] = u[8] & 0x3F // byte 8 without variant bits
	copy(msg[3:], u[9:]) // bytes 9-15
	return msg
}

// setVersion sets the version number of the UUID.
func setVersion(u *UUID, ver byte) {
	u[6] = (u[6] & 0x0F) | ((ver & 0x0F) << 4)
}

// setVariantRFC4122 sets the variant bits to RFC 4122 (10xxxxxx).
func setVariantRFC4122(u *UUID) {
	u[8] = (u[8] & 0x3F) | 0x80
}
