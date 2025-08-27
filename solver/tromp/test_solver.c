// Simple test program for solver
#include "cuckoo_lean.h"
#include <stdio.h>
#include <string.h>

int main() {
    printf("Testing Cuckoo solver...\n");
    
    solver_ctx ctx;
    cuckoo_init(&ctx);
    
    printf("Initialized context:\n");
    printf("  nthreads: %u\n", ctx.nthreads);
    printf("  nonce_range: %u\n", ctx.nonce_range);
    
    // Set test header
    uint8_t header[80];
    memset(header, 0, 80);
    strcpy((char*)header, "TEST HEADER");
    
    printf("Setting header...\n");
    cuckoo_setheader(&ctx, header, 80);
    
    printf("Header set successfully\n");
    printf("Attempting to solve (nonce 0, range 1)...\n");
    
    ctx.nonce = 0;
    ctx.nonce_range = 1;
    ctx.nthreads = 1;
    
    int solutions = cuckoo_solve(&ctx);
    printf("Solutions found: %d\n", solutions);
    
    return 0;
}
