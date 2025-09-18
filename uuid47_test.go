package uuid47

import (
	"testing"

	"github.com/dchest/siphash"
)

// Test helper functions

// version returns the version number of the UUID.
func version(u UUID) int {
	return int((u[6] >> 4) & 0x0F)
}

// craftV7 creates a UUIDv7 with the specified timestamp and random bits.
// This is for testing to match the C implementation's craft_v7 function.
func craftV7(tsMs48 uint64, randA12 uint16, randB62 uint64) UUID {
	var u UUID

	// Write 48-bit timestamp
	wr48be(u[:6], tsMs48&0x0000FFFFFFFFFFFF)

	// Set version to 7
	setVersion(&u, 7)

	// Set rand_a (12 bits)
	u[6] = (u[6] & 0xF0) | byte((randA12>>8)&0x0F)
	u[7] = byte(randA12 & 0xFF)

	// Set variant
	setVariantRFC4122(&u)

	// Set rand_b (62 bits across bytes 8-15)
	// First 6 bits go in byte 8 (lower 6 bits)
	u[8] = (u[8] & 0xC0) | byte((randB62>>56)&0x3F)
	// Remaining 56 bits go in bytes 9-15
	for i := 0; i < 7; i++ {
		u[9+i] = byte(randB62 >> (48 - i*8))
	}

	return u
}

// Test vectors from C implementation's tests.c
func TestNewRandomKey(t *testing.T) {
	// Test that NewRandomKey generates different keys
	key1, err := NewRandomKey()
	if err != nil {
		t.Fatal(err)
	}

	key2, err := NewRandomKey()
	if err != nil {
		t.Fatal(err)
	}

	if key1 == key2 {
		t.Error("NewRandomKey generated identical keys")
	}

	// Test that keys work with Encode/Decode
	u, _ := Parse("018f2d9f-9a2a-7def-8c3f-7b1a2c4d5e6f")
	encoded := Encode(u, key1)
	decoded := Decode(encoded, key1)

	if decoded != u {
		t.Error("Random key failed roundtrip test")
	}
}

func TestSipHashVectors(t *testing.T) {
	// From test_siphash_switch_and_vectors_subset in tests.c
	k0 := uint64(0x0706050403020100)
	k1 := uint64(0x0f0e0d0c0b0a0908)

	// Note: The C test vectors are stored as little-endian bytes
	// We need to interpret them as little-endian uint64 values
	vectors := []struct {
		length   int
		expected uint64
	}{
		{0, 0x726fdb47dd0e0e31},  // {0x31, 0x0e, 0x0e, 0xdd, 0x47, 0xdb, 0x6f, 0x72} in LE
		{1, 0x74f839c593dc67fd},  // {0xfd, 0x67, 0xdc, 0x93, 0xc5, 0x39, 0xf8, 0x74} in LE
		{2, 0x0d6c8009d9a94f5a},  // {0x5a, 0x4f, 0xa9, 0xd9, 0x09, 0x80, 0x6c, 0x0d} in LE
		{3, 0x85676696d7fb7e2d},  // {0x2d, 0x7e, 0xfb, 0xd7, 0x96, 0x66, 0x67, 0x85} in LE
		{4, 0xcf2794e0277187b7},  // {0xb7, 0x87, 0x71, 0x27, 0xe0, 0x94, 0x27, 0xcf} in LE
		{5, 0x18765564cd99a68d},  // {0x8d, 0xa6, 0x99, 0xcd, 0x64, 0x55, 0x76, 0x18} in LE
		{6, 0xcbc9466e58fee3ce},  // {0xce, 0xe3, 0xfe, 0x58, 0x6e, 0x46, 0xc9, 0xcb} in LE
		{7, 0xab0200f58b01d137},  // {0x37, 0xd1, 0x01, 0x8b, 0xf5, 0x00, 0x02, 0xab} in LE
		{8, 0x93f5f5799a932462},  // {0x62, 0x24, 0x93, 0x9a, 0x79, 0xf5, 0xf5, 0x93} in LE
		{9, 0x9e0082df0ba9e4b0},  // {0xb0, 0xe4, 0xa9, 0x0b, 0xdf, 0x82, 0x00, 0x9e} in LE
		{10, 0x7a5dbbc594ddb9f3}, // {0xf3, 0xb9, 0xdd, 0x94, 0xc5, 0xbb, 0x5d, 0x7a} in LE
		{11, 0xf4b32f46226bada7}, // {0xa7, 0xad, 0x6b, 0x22, 0x46, 0x2f, 0xb3, 0xf4} in LE
		{12, 0x751e8fbc860ee5fb}, // {0xfb, 0xe5, 0x0e, 0x86, 0xbc, 0x8f, 0x1e, 0x75} in LE
	}

	msg := make([]byte, 64)
	for i := range msg {
		msg[i] = byte(i)
	}

	for _, v := range vectors {
		got := siphash.Hash(k0, k1, msg[:v.length])
		if got != v.expected {
			t.Errorf("SipHash mismatch for length %d: got %016x, want %016x",
				v.length, got, v.expected)
		}
	}
}

func TestRd48BeWr48Be(t *testing.T) {
	// Test from test_rd_wr_48 in tests.c
	buf := make([]byte, 6)
	v := uint64(0x0123456789AB) & 0x0000FFFFFFFFFFFF

	wr48be(buf, v)
	r := rd48be(buf)

	if r != v {
		t.Errorf("rd48be/wr48be roundtrip failed: got %016x, want %016x", r, v)
	}
}

func TestUUIDParseFormatRoundtrip(t *testing.T) {
	// Test from test_uuid_parse_format_roundtrip in tests.c
	s := "00000000-0000-7000-8000-000000000000"

	u, err := Parse(s)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if version(u) != 7 {
		t.Errorf("Version mismatch: got %d, want 7", version(u))
	}

	out := u.String()
	u2, err := Parse(out)
	if err != nil {
		t.Fatalf("Parse roundtrip failed: %v", err)
	}

	if u != u2 {
		t.Errorf("Parse/format roundtrip mismatch: %v != %v", u, u2)
	}

	// Test invalid UUID
	bad := "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz"
	_, err = Parse(bad)
	if err == nil {
		t.Error("Parse should have failed for invalid UUID")
	}
}

func TestVersionVariant(t *testing.T) {
	// Test from test_version_variant in tests.c
	var u UUID
	setVersion(&u, 7)
	if version(u) != 7 {
		t.Errorf("SetVersion failed: got %d, want 7", version(u))
	}

	setVariantRFC4122(&u)
	if (u[8] & 0xC0) != 0x80 {
		t.Errorf("SetVariantRFC4122 failed: got %02x, want 10xxxxxx", u[8])
	}
}

func TestEncodeDecodeRoundtrip(t *testing.T) {
	// Test from test_encode_decode_roundtrip in tests.c
	key := Key{K0: 0x0123456789abcdef, K1: 0xfedcba9876543210}

	for i := 0; i < 16; i++ {
		// Match the C test exactly
		ts := uint64((0x100000*uint64(i) + 123))
		ra := uint16((0x0AAA ^ uint32(i*7)) & 0x0FFF)
		rb := (uint64(0x0123456789ABCDEF) ^ (0x1111111111111111 * uint64(i))) & ((1 << 62) - 1)

		u7 := craftV7(ts, ra, rb)

		// Encode to v4 facade
		facade := Encode(u7, key)

		// Check version is 4
		if version(facade) != 4 {
			t.Errorf("Facade version should be 4, got %d", version(facade))
		}

		// Check variant bits
		if (facade[8] & 0xC0) != 0x80 {
			t.Errorf("Facade variant bits incorrect: got %02x", facade[8])
		}

		// Decode back to v7
		back := Decode(facade, key)

		// Should match original
		if back != u7 {
			t.Errorf("Roundtrip failed for iteration %d:\nOriginal: %v\nDecoded:  %v",
				i, u7, back)
		}

		// Test with wrong key (should not match)
		wrongKey := Key{K0: key.K0 ^ 0xdeadbeef, K1: key.K1 ^ 0x1337}
		bad := Decode(facade, wrongKey)
		if bad == u7 {
			t.Error("Decode with wrong key should not match original")
		}
	}
}

// TestExactCCompatibility verifies our implementation matches C exactly
func TestExactCCompatibility(t *testing.T) {
	// These test vectors were generated from the C implementation
	testCases := []struct {
		name       string
		key        Key
		inputV7    string
		expectedV4 string
	}{
		{
			name:       "C demo.c example",
			key:        Key{K0: 0x0123456789abcdef, K1: 0xfedcba9876543210},
			inputV7:    "018f2d9f-9a2a-7def-8c3f-7b1a2c4d5e6f",
			expectedV4: "2463c780-7fca-4def-8c3f-7b1a2c4d5e6f",
		},
		{
			name:       "All zeros timestamp",
			key:        Key{K0: 0x0123456789abcdef, K1: 0xfedcba9876543210},
			inputV7:    "00000000-0000-7000-8000-000000000000",
			expectedV4: "22d97126-9609-4000-8000-000000000000",
		},
		{
			name:       "Test vector 0 from C roundtrip test",
			key:        Key{K0: 0x0123456789abcdef, K1: 0xfedcba9876543210},
			inputV7:    "00000000-007b-7aaa-8123-456789abcdef",
			expectedV4: "b108050e-46b6-4aaa-8123-456789abcdef",
		},
		{
			name:       "Test vector 1 from C roundtrip test",
			key:        Key{K0: 0x0123456789abcdef, K1: 0xfedcba9876543210},
			inputV7:    "00000010-007b-7aad-9032-547698badcfe",
			expectedV4: "bc75bd50-97ef-4aad-9032-547698badcfe",
		},
		{
			name:       "Test vector 2 from C roundtrip test",
			key:        Key{K0: 0x0123456789abcdef, K1: 0xfedcba9876543210},
			inputV7:    "00000020-007b-7aa4-a301-6745ab89efcd",
			expectedV4: "a3e09c87-bf85-4aa4-a301-6745ab89efcd",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v7, err := Parse(tc.inputV7)
			if err != nil {
				t.Fatalf("Failed to parse input v7: %v", err)
			}

			facade := Encode(v7, tc.key)
			gotV4 := facade.String()

			if gotV4 != tc.expectedV4 {
				t.Errorf("Encode mismatch:\nGot:      %s\nExpected: %s", gotV4, tc.expectedV4)
			}

			// Verify decode works
			decoded := Decode(facade, tc.key)
			if decoded.String() != tc.inputV7 {
				t.Errorf("Decode mismatch:\nGot:      %s\nExpected: %s",
					decoded.String(), tc.inputV7)
			}
		})
	}
}

func TestCraftV7(t *testing.T) {
	// Test the CraftV7 function matches C's craft_v7
	tsMs48 := uint64(0x0123456789AB)
	randA12 := uint16(0x0ABC)
	randB62 := uint64(0x0FEDCBA987654321) & ((1 << 62) - 1)

	u := craftV7(tsMs48, randA12, randB62)

	// Check timestamp
	gotTs := rd48be(u[:6])
	if gotTs != (tsMs48 & 0x0000FFFFFFFFFFFF) {
		t.Errorf("Timestamp mismatch: got %012x, want %012x",
			gotTs, tsMs48&0x0000FFFFFFFFFFFF)
	}

	// Check version
	if version(u) != 7 {
		t.Errorf("Version should be 7, got %d", version(u))
	}

	// Check rand_a (12 bits)
	gotRandA := (uint16(u[6]&0x0F) << 8) | uint16(u[7])
	if gotRandA != (randA12 & 0x0FFF) {
		t.Errorf("rand_a mismatch: got %03x, want %03x",
			gotRandA, randA12&0x0FFF)
	}

	// Check variant
	if (u[8] & 0xC0) != 0x80 {
		t.Errorf("Variant bits incorrect: got %02x", u[8])
	}

	// Check rand_b (62 bits, stored across bytes 8-15)
	// First 6 bits from byte 8, then 56 bits from bytes 9-15
	gotRandB := uint64(u[8]&0x3F) << 56
	for i := 0; i < 7; i++ {
		gotRandB |= uint64(u[9+i]) << (48 - i*8)
	}

	if gotRandB != (randB62 & ((1 << 62) - 1)) {
		t.Errorf("rand_b mismatch: got %015x, want %015x",
			gotRandB, randB62&((1<<62)-1))
	}
}

func TestBuildSipInput(t *testing.T) {
	// Test that buildSipInputFromV7 matches C's build_sip_input_from_v7
	u, err := Parse("018f2d9f-9a2a-7def-8c3f-7b1a2c4d5e6f")
	if err != nil {
		t.Fatal(err)
	}
	sipInput := buildSipInputFromV7(u)

	// The SIP input should be:
	// [low-nibble of b6][b7][b8&0x3F][b9..b15]
	expected := [10]byte{
		u[6] & 0x0F, // 0xF (from 0xef)
		u[7],        // 0xef
		u[8] & 0x3F, // 0x0c (from 0x8c)
		u[9],        // 0x3f
		u[10],       // 0x7b
		u[11],       // 0x1a
		u[12],       // 0x2c
		u[13],       // 0x4d
		u[14],       // 0x5e
		u[15],       // 0x6f
	}

	if sipInput != expected {
		t.Errorf("buildSipInputFromV7 mismatch:\nGot:      %x\nExpected: %x",
			sipInput, expected)
	}
}

// Benchmarks
func BenchmarkEncode(b *testing.B) {
	u7, _ := Parse("018f2d9f-9a2a-7def-8c3f-7b1a2c4d5e6f")
	key := Key{K0: 0x0123456789abcdef, K1: 0xfedcba9876543210}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Encode(u7, key)
	}
}

func BenchmarkDecode(b *testing.B) {
	u7, _ := Parse("018f2d9f-9a2a-7def-8c3f-7b1a2c4d5e6f")
	key := Key{K0: 0x0123456789abcdef, K1: 0xfedcba9876543210}
	facade := Encode(u7, key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Decode(facade, key)
	}
}

func BenchmarkRoundtrip(b *testing.B) {
	u7, _ := Parse("018f2d9f-9a2a-7def-8c3f-7b1a2c4d5e6f")
	key := Key{K0: 0x0123456789abcdef, K1: 0xfedcba9876543210}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		facade := Encode(u7, key)
		_ = Decode(facade, key)
	}
}

func BenchmarkParse(b *testing.B) {
	s := "018f2d9f-9a2a-7def-8c3f-7b1a2c4d5e6f"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(s)
	}
}

func BenchmarkString(b *testing.B) {
	u, _ := Parse("018f2d9f-9a2a-7def-8c3f-7b1a2c4d5e6f")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = u.String()
	}
}
