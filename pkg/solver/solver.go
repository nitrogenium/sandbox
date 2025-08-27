// Package solver provides Go interface to Cuckoo Cycle solver
package solver

/*
#cgo CFLAGS: -I../../solver/tromp -I../../solver/tromp/cuckoo-orig/src
#cgo CXXFLAGS: -O3 -march=native -std=c++14 -pthread
#cgo LDFLAGS: -L../../solver/tromp -lcuckoo_lean -lstdc++ -lpthread

#include "cuckoo_lean.h"
#include <stdlib.h>
// Allocate/free helpers with external linkage for cgo
solver_ctx* alloc_solver_ctx() { return (solver_ctx*)malloc(sizeof(solver_ctx)); }
void free_solver_ctx(solver_ctx* p) { if (p) free(p); }
// Forward declaration and wrapper for abort to satisfy cgo
void cuckoo_abort(solver_ctx* ctx);
void go_cuckoo_abort(solver_ctx* ctx) { cuckoo_abort(ctx); }
*/
import "C"
import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"unsafe"
)

const (
	ProofSize = 42
	MaxSols   = 8
	EdgeBits  = 23
)

// Solution represents a Cuckoo Cycle solution
type Solution struct {
	Nonce []uint32
}

// Solver wraps the C++ Cuckoo solver
type Solver struct {
	ctx      C.solver_ctx
	nthreads int
}

// NewSolver creates a new Cuckoo solver with specified threads
func NewSolver(nthreads int) *Solver {
	s := &Solver{nthreads: nthreads}
	// Initialize C context in-place
	C.cuckoo_init(&s.ctx)
	s.ctx.nthreads = C.uint32_t(nthreads)
	return s
}

// SetHeader sets the header data for mining
func (s *Solver) SetHeader(header []byte) {
	if len(header) > 80 {
		header = header[:80]
	}
	if len(header) == 0 {
		panic("SetHeader: empty header")
	}
	C.cuckoo_setheader(&s.ctx, (*C.uint8_t)(unsafe.Pointer(&header[0])), C.uint32_t(len(header)))
}

// Solve searches for Cuckoo cycles in the given nonce range
func (s *Solver) Solve(baseNonce uint32, nonceRange uint32) []Solution {
	s.ctx.nonce = C.uint32_t(baseNonce)
	s.ctx.nonce_range = C.uint32_t(nonceRange)

	nsols := int(C.cuckoo_solve(&s.ctx))

	solutions := make([]Solution, 0, nsols)

	// Access C array through pointer arithmetic
	proofsPtr := (*[MaxSols]C.proof_t)(unsafe.Pointer(&s.ctx.proofs[0]))

	for i := 0; i < nsols; i++ {
		sol := Solution{
			Nonce: make([]uint32, ProofSize),
		}
		// Access nonce array in each proof
		noncePtr := (*[ProofSize]C.uint32_t)(unsafe.Pointer(&proofsPtr[i].nonce[0]))
		for j := 0; j < ProofSize; j++ {
			sol.Nonce[j] = uint32(noncePtr[j])
		}
		solutions = append(solutions, sol)
	}

	return solutions
}

// Cancel requests to abort a running solve (best effort).
func (s *Solver) Cancel() {
	C.go_cuckoo_abort(&s.ctx)
}

// Verify checks if a solution is valid
func Verify(header []byte, nonce uint32, proof []uint32) bool {
	if len(proof) != ProofSize {
		return false
	}

	cProof := make([]C.uint32_t, ProofSize)
	for i, p := range proof {
		cProof[i] = C.uint32_t(p)
	}

	result := C.cuckoo_verify(
		(*C.uint8_t)(unsafe.Pointer(&header[0])),
		C.uint32_t(len(header)),
		C.uint32_t(nonce),
		(*C.uint32_t)(unsafe.Pointer(&cProof[0])),
	)

	return result != 0
}

// HashSolution computes SHA256d hash of the solution for difficulty check
func HashSolution(header []byte, nonce uint32, solution []uint32) [32]byte {
	// Build data: header + nonce + solution
	data := make([]byte, len(header)+4+len(solution)*4)
	copy(data, header)
	binary.LittleEndian.PutUint32(data[len(header):], nonce)

	offset := len(header) + 4
	for _, s := range solution {
		binary.LittleEndian.PutUint32(data[offset:], s)
		offset += 4
	}

	// SHA256d
	h1 := sha256.Sum256(data)
	h2 := sha256.Sum256(h1[:])
	return h2
}

// CheckDifficulty checks if solution hash meets target difficulty
func CheckDifficulty(hash [32]byte, target []byte) bool {
	// Compare as big-endian (Bitcoin style)
	for i := 0; i < len(target) && i < 32; i++ {
		if hash[31-i] > target[31-i] {
			return false
		}
		if hash[31-i] < target[31-i] {
			return true
		}
	}
	return true
}

// GetStats returns solver statistics
func (s *Solver) GetStats() string {
	return fmt.Sprintf("Solver: %d threads, EdgeBits: %d, ProofSize: %d",
		s.nthreads, EdgeBits, ProofSize)
}

// Close releases C-side resources for the solver context.
func (s *Solver) Close() {}
