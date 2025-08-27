// Cuckoo Cycle lean solver wrapper
// Adapts John Tromp's implementation for Go integration

#include "cuckoo_lean.h"
#include <string.h>
#include <stdlib.h>
#include <pthread.h>

// Include Tromp's implementation with our parameters
#define HEADERLEN 80
#include "cuckoo-orig/src/cuckoo/lean.hpp"
#include "cuckoo-orig/src/crypto/blake2b-ref.c"

// Thread context for parallel solving
struct thread_ctx {
    int id;
    cuckoo_ctx* ctx;
    pthread_t thread;
};

// Wrapper context that includes Tromp's context
struct internal_ctx {
    cuckoo_ctx* tromp_ctx;
    thread_ctx* threads;
    solver_ctx* api_ctx;
};

// Global storage for internal contexts
static internal_ctx* g_internal_ctx = NULL;

extern "C" {

void cuckoo_init(solver_ctx* ctx) {
    memset(ctx, 0, sizeof(solver_ctx));
    ctx->nthreads = 1;
    ctx->nonce_range = 1;
}

void cuckoo_setheader(solver_ctx* ctx, const uint8_t* header, uint32_t len) {
    if (len > 80) len = 80;
    memcpy(ctx->header, header, len);
    ctx->header_len = len;
}

int cuckoo_solve(solver_ctx* ctx) {
    // Create internal context
    internal_ctx* ictx = new internal_ctx;
    ictx->api_ctx = ctx;
    
    // Initialize Tromp's context
    int ntrims = 2 + (PART_BITS+3)*(PART_BITS+4);
    ictx->tromp_ctx = new cuckoo_ctx(ctx->nthreads, ntrims, MAXSOLS);
    
    // Allocate thread contexts
    ictx->threads = new thread_ctx[ctx->nthreads];
    
    ctx->solutions = 0;
    
    // Search through nonce range
    for (uint32_t r = 0; r < ctx->nonce_range && ctx->solutions < MAXSOLS; r++) {
        // Set header and nonce
        ictx->tromp_ctx->setheadernonce((char*)ctx->header, ctx->header_len, ctx->nonce + r);
        ictx->tromp_ctx->barry.clear();
        
        // Launch threads
        for (int t = 0; t < ctx->nthreads; t++) {
            ictx->threads[t].id = t;
            ictx->threads[t].ctx = ictx->tromp_ctx;
            pthread_create(&ictx->threads[t].thread, NULL, worker, (void*)&ictx->threads[t]);
        }
        
        // Wait for threads
        for (int t = 0; t < ctx->nthreads; t++) {
            pthread_join(ictx->threads[t].thread, NULL);
        }
        
        // Copy solutions
        for (unsigned s = 0; s < ictx->tromp_ctx->nsols && ctx->solutions < MAXSOLS; s++) {
            for (int i = 0; i < PROOFSIZE; i++) {
                ctx->proofs[ctx->solutions].nonce[i] = ictx->tromp_ctx->sols[s][i];
            }
            ctx->solutions++;
        }
    }
    
    // Cleanup
    delete[] ictx->threads;
    delete ictx->tromp_ctx;
    delete ictx;
    
    return ctx->solutions;
}

int cuckoo_verify(const uint8_t* header, uint32_t header_len, uint32_t nonce, const uint32_t* proof) {
    siphash_keys keys;
    char headernonce[88];
    
    // Prepare header with nonce
    memcpy(headernonce, header, header_len);
    memcpy(headernonce + header_len, &nonce, sizeof(nonce));
    
    // Generate siphash keys
    blake2b((void*)&keys, sizeof(keys), headernonce, header_len + sizeof(nonce), 0, 0);
    
    // Verify the proof
    u32 uvs[2*PROOFSIZE];
    u32 xor0 = 0, xor1 = 0;
    
    for (u32 n = 0; n < PROOFSIZE; n++) {
        if (n > 0 && proof[n] <= proof[n-1])
            return 0;
        u32 edge = sipnode(&keys, proof[n], 0);
        u32 node0 = edge & EDGEMASK;
        u32 node1 = (edge >> 32) & EDGEMASK;
        uvs[2*n] = node0;
        uvs[2*n+1] = node1;
        xor0 ^= node0;
        xor1 ^= node1;
    }
    
    if (xor0 | xor1)
        return 0;
    
    // Check cycle
    u32 n = 0, i = 0;
    do {
        u32 j = i;
        for (u32 k = 0; k < 2*PROOFSIZE; k += 2) {
            if (k != i && uvs[k] == uvs[i]) {
                if (j != i)
                    return 0;
                j = k;
            }
        }
        if (j == i)
            return 0;
        i = j^1;
        n++;
    } while (i != 0);
    
    return n == PROOFSIZE;
}

void cuckoo_sha256d(const uint8_t* data, size_t len, uint8_t* hash) {
    // Use blake2b for now, should be replaced with actual SHA256d
    blake2b(hash, 32, data, len, NULL, 0);
    blake2b(hash, 32, hash, 32, NULL, 0);
}

} // extern "C"
