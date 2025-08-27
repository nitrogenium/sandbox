#include "cuckoo_lean.h"
#include <stdio.h>
#include <string.h>
#include <stdint.h>

int main() {
    printf("=== Isolated Cuckoo Solver Test ===\n");

    solver_ctx ctx;
    cuckoo_init(&ctx);

    printf("nthreads=%u nonce_range=%u\n", ctx.nthreads, ctx.nonce_range);

    // Prepare 80-byte header filled with zeros except a marker
    uint8_t header[80];
    memset(header, 0, sizeof(header));
    memcpy(header, "TEST-HEADER", 11);

    cuckoo_setheader(&ctx, header, 80);

    ctx.nthreads = 1;
    ctx.nonce = 0;
    ctx.nonce_range = 1;

    int sols = cuckoo_solve(&ctx);
    printf("cuckoo_solve returned %d\n", sols);

    // Try abort test quickly (should be no-op after return)
    cuckoo_abort(&ctx);

    return 0;
}
