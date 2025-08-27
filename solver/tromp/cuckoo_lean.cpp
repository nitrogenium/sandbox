// Cuckoo Cycle lean solver wrapper
// Adapts John Tromp's implementation for Go integration

#include "cuckoo_lean.h"
#include <string.h>
#include <stdlib.h>
#include <pthread.h>
#include <stdio.h>

// Include Tromp's implementation with our parameters
#define HEADERLEN 80
#include "cuckoo-orig/src/cuckoo/lean.hpp"
#include "cuckoo-orig/src/crypto/blake2b-ref.c"

// Forward declaration from lean.hpp
void *worker(void *vp);

// Wrapper context that includes Tromp's context
struct internal_ctx {
    cuckoo_ctx* tromp_ctx;
    thread_ctx* threads;
    solver_ctx* api_ctx;
};

// Global storage for internal contexts (removed - not needed)

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
    fprintf(stderr, "[DEBUG] cuckoo_solve: starting, nthreads=%u, nonce=%u, range=%u\n", 
            ctx->nthreads, ctx->nonce, ctx->nonce_range);
    
    // Create internal context
    internal_ctx* ictx = new internal_ctx;
    ictx->api_ctx = ctx;
    
    fprintf(stderr, "[DEBUG] cuckoo_solve: calculating ntrims\n");
    // Initialize Tromp's context
    int ntrims = 2 + (PART_BITS+3)*(PART_BITS+4);
    
    fprintf(stderr, "[DEBUG] cuckoo_solve: creating cuckoo_ctx with nthreads=%u, ntrims=%d\n", 
            ctx->nthreads, ntrims);
    
    try {
        ictx->tromp_ctx = new cuckoo_ctx(ctx->nthreads, ntrims, MAXSOLS);
        fprintf(stderr, "[DEBUG] cuckoo_solve: cuckoo_ctx created successfully\n");
    } catch (std::bad_alloc& e) {
        fprintf(stderr, "[ERROR] cuckoo_solve: failed to allocate cuckoo_ctx: %s\n", e.what());
        delete ictx;
        return 0;
    } catch (...) {
        fprintf(stderr, "[ERROR] cuckoo_solve: unknown error creating cuckoo_ctx\n");
        delete ictx;
        return 0;
    }
    
    fprintf(stderr, "[DEBUG] cuckoo_solve: allocating thread contexts\n");
    // Allocate thread contexts
    ictx->threads = new thread_ctx[ctx->nthreads];
    
    ctx->solutions = 0;
    
    // Search through nonce range
    for (uint32_t r = 0; r < ctx->nonce_range && ctx->solutions < MAXSOLS; r++) {
        // Set header and nonce
        ictx->tromp_ctx->setheadernonce((char*)ctx->header, ctx->header_len, ctx->nonce + r);
        ictx->tromp_ctx->barry.clear();
        
        fprintf(stderr, "[DEBUG] cuckoo_solve: launching %u threads for nonce %u\n", ctx->nthreads, ctx->nonce + r);
        // Launch threads
        for (uint32_t t = 0; t < ctx->nthreads; t++) {
            ictx->threads[t].id = t;
            ictx->threads[t].ctx = ictx->tromp_ctx;
            fprintf(stderr, "[DEBUG] cuckoo_solve: creating thread %u\n", t);
            int err = pthread_create(&ictx->threads[t].thread, NULL, worker, (void*)&ictx->threads[t]);
            if (err) {
                fprintf(stderr, "[ERROR] cuckoo_solve: failed to create thread %u: %d\n", t, err);
                // Handle thread creation error
                delete[] ictx->threads;
                delete ictx->tromp_ctx;
                delete ictx;
                return 0;
            }
        }
        
        // Wait for threads
        for (uint32_t t = 0; t < ctx->nthreads; t++) {
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
    setheader(headernonce, header_len + sizeof(nonce), &keys);
    
    // Verify the proof
    u32 uvs[2*PROOFSIZE];
    u32 xor0 = 0, xor1 = 0;
    
    for (u32 n = 0; n < PROOFSIZE; n++) {
        if (n > 0 && proof[n] <= proof[n-1])
            return 0;
        u32 node0 = sipnode(&keys, proof[n], 0);
        u32 node1 = sipnode(&keys, proof[n], 1);
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
