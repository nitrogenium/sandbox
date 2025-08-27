// Simplified Cuckoo Cycle lean solver interface for Go
// Based on John Tromp's implementation

#ifndef CUCKOO_LEAN_H
#define CUCKOO_LEAN_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <stddef.h>

// Parameters matching Java miner
#define EDGEBITS 23
#define PROOFSIZE 42
#define MAXSOLS 8

typedef struct {
    uint32_t nonce[PROOFSIZE];
} proof_t;

typedef struct {
    uint8_t header[80];    // Header data
    uint32_t header_len;   // Actual header length
    uint32_t nonce;        // Base nonce
    uint32_t nonce_range;  // Nonce range to search
    uint32_t nthreads;     // Number of threads
    uint32_t solutions;    // Number of solutions found
    proof_t proofs[MAXSOLS]; // Found solutions
} solver_ctx;

// Initialize solver context
void cuckoo_init(solver_ctx* ctx);

// Set header for mining
void cuckoo_setheader(solver_ctx* ctx, const uint8_t* header, uint32_t len);

// Find cycles in nonce range
int cuckoo_solve(solver_ctx* ctx);

// Verify a solution
int cuckoo_verify(const uint8_t* header, uint32_t header_len, uint32_t nonce, const uint32_t* proof);

// Hash function for target checking (SHA256d)
void cuckoo_sha256d(const uint8_t* data, size_t len, uint8_t* hash);

#ifdef __cplusplus
}
#endif

#endif // CUCKOO_LEAN_H
