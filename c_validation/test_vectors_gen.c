// Generate test vectors from the C implementation
#include <stdio.h>
#include <string.h>

// Download uuidv47.h from https://raw.githubusercontent.com/stateless-me/uuidv47/main/uuidv47.h
// and place it in the same directory
#include "uuidv47.h"

int main(void) {
    uuidv47_key_t key = {.k0 = 0x0123456789abcdefULL, .k1 = 0xfedcba9876543210ULL};

    // Test case 1: The demo.c example
    {
        uuid128_t v7, facade, back;
        const char *input = "018f2d9f-9a2a-7def-8c3f-7b1a2c4d5e6f";
        uuid_parse(input, &v7);

        facade = uuidv47_encode_v4facade(v7, key);
        back = uuidv47_decode_v4facade(facade, key);

        char v7_str[37], facade_str[37], back_str[37];
        uuid_format(&v7, v7_str);
        uuid_format(&facade, facade_str);
        uuid_format(&back, back_str);

        printf("Test case: demo.c example\n");
        printf("Input v7:  %s\n", v7_str);
        printf("V4 facade: %s\n", facade_str);
        printf("Decoded:   %s\n", back_str);
        printf("\n");
    }

    // Test case 2: All zeros timestamp
    {
        uuid128_t v7, facade;
        const char *input = "00000000-0000-7000-8000-000000000000";
        uuid_parse(input, &v7);

        facade = uuidv47_encode_v4facade(v7, key);

        char v7_str[37], facade_str[37];
        uuid_format(&v7, v7_str);
        uuid_format(&facade, facade_str);

        printf("Test case: All zeros timestamp\n");
        printf("Input v7:  %s\n", v7_str);
        printf("V4 facade: %s\n", facade_str);
        printf("\n");
    }

    // Generate the test vectors from test_encode_decode_roundtrip
    printf("Test vectors from test_encode_decode_roundtrip:\n");
    for (int i = 0; i < 3; i++) {  // Just first 3 for verification
        uuid128_t u7;
        uint64_t ts = (uint64_t)((0x100000ULL * (uint64_t)i) + 123ULL);
        uint16_t ra = (uint16_t)((0x0AAA ^ (uint32_t)(i * 7)) & 0x0FFF);
        uint64_t rb = (0x0123456789ABCDEFULL ^ (0x1111111111111111ULL * (uint64_t)i)) & ((1ULL << 62) - 1);

        // Craft the v7
        memset(&u7, 0, sizeof(u7));
        wr48be(&u7.b[0], ts & 0x0000FFFFFFFFFFFFULL);
        set_version(&u7, 7);
        u7.b[6] = (uint8_t)((u7.b[6] & 0xF0) | ((ra >> 8) & 0x0F));
        u7.b[7] = (uint8_t)(ra & 0xFF);
        set_variant_rfc4122(&u7);
        u7.b[8] = (uint8_t)((u7.b[8] & 0xC0) | ((rb >> 56) & 0x3F));
        for (int j = 0; j < 7; j++) {
            u7.b[9 + j] = (uint8_t)(rb >> (48 - j * 8));
        }

        uuid128_t facade = uuidv47_encode_v4facade(u7, key);

        char v7_str[37], facade_str[37];
        uuid_format(&u7, v7_str);
        uuid_format(&facade, facade_str);

        printf("  [%d] v7: %s -> v4: %s\n", i, v7_str, facade_str);
    }

    return 0;
}