// Simple test to verify Tromp's Cuckoo solver
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define HEADERLEN 80
#include "cuckoo-orig/src/cuckoo/lean.hpp"
#include "cuckoo-orig/src/crypto/blake2b-ref.c"

int main() {
    printf("Testing Tromp's Cuckoo solver directly...\n");
    
    // Test parameters
    const int nthreads = 1;
    const int ntrims = 2 + (PART_BITS+3)*(PART_BITS+4);
    const int maxsols = 8;
    
    printf("Parameters: nthreads=%d, ntrims=%d, PART_BITS=%d\n", nthreads, ntrims, PART_BITS);
    printf("Memory sizes: NEDGES=%d, shrinkingset needs %ld KB\n", NEDGES, (long)(NEDGES/8/1024));
    
    // Create context
    printf("Creating cuckoo_ctx...\n");
    cuckoo_ctx* ctx = nullptr;
    try {
        ctx = new cuckoo_ctx(nthreads, ntrims, maxsols);
        printf("✓ cuckoo_ctx created\n");
    } catch (std::bad_alloc& e) {
        printf("✗ Failed to allocate cuckoo_ctx: %s\n", e.what());
        return 1;
    } catch (...) {
        printf("✗ Unknown error creating cuckoo_ctx\n");
        return 1;
    }
    
    // Test header
    char header[88] = {0};
    memset(header, 0x41, 80); // Fill with 'A'
    uint32_t nonce = 42;
    memcpy(header + 80, &nonce, sizeof(nonce));
    
    printf("Setting header...\n");
    ctx->setheadernonce(header, 84, nonce);
    printf("✓ Header set\n");
    
    // Create thread context
    thread_ctx tc;
    tc.id = 0;
    tc.ctx = ctx;
    
    printf("Running worker...\n");
    worker(&tc);
    printf("✓ Worker completed\n");
    
    printf("Solutions found: %d\n", ctx->nsols.load());
    
    // Cleanup
    delete ctx;
    printf("✓ Cleanup complete\n");
    
    return 0;
}
