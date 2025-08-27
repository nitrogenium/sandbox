// Test C API wrapper
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "cuckoo_lean.h"

int main() {
    printf("=== Testing C API Wrapper ===\n");
    
    // Create solver context
    solver_ctx ctx;
    printf("Initializing solver context...\n");
    cuckoo_init(&ctx);
    printf("✓ Initialized\n");
    
    // Set parameters
    ctx.nthreads = 1;
    ctx.nonce = 0;
    ctx.nonce_range = 1;
    
    // Set header
    uint8_t header[80];
    memset(header, 0x41, 80); // Fill with 'A'
    printf("Setting header (80 bytes of 'A')...\n");
    cuckoo_setheader(&ctx, header, 80);
    printf("✓ Header set\n");
    
    // Try to solve
    printf("Calling cuckoo_solve with:\n");
    printf("  nthreads: %u\n", ctx.nthreads);
    printf("  nonce: %u\n", ctx.nonce); 
    printf("  nonce_range: %u\n", ctx.nonce_range);
    
    int solutions = cuckoo_solve(&ctx);
    
    printf("✓ cuckoo_solve returned: %d solutions\n", solutions);
    
    // Test verify (even if no solution)
    uint32_t proof[42] = {0};
    printf("Testing cuckoo_verify...\n");
    int valid = cuckoo_verify(header, 80, 0, proof);
    printf("✓ cuckoo_verify returned: %d\n", valid);
    
    printf("=== Test Complete ===\n");
    return 0;
}
