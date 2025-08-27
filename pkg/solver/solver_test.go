package solver

import (
	"encoding/hex"
	"testing"
)

func TestSolver(t *testing.T) {
	// Test header (80 bytes)
	headerHex := "0100000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"00000000ffff001d00000000"

	header, err := hex.DecodeString(headerHex)
	if err != nil {
		t.Fatal(err)
	}

	// Create solver
	s := NewSolver(1)
	s.SetHeader(header)

	// Try to find solution
	solutions := s.Solve(0, 1000)

	t.Logf("Found %d solutions", len(solutions))
	for i, sol := range solutions {
		t.Logf("Solution %d: %v", i, sol.Nonce)

		// Verify solution
		if !Verify(header, 0, sol.Nonce) {
			t.Errorf("Solution %d failed verification", i)
		}
	}
}

func BenchmarkSolver(b *testing.B) {
	headerHex := "0100000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"00000000ffff001d00000000"

	header, _ := hex.DecodeString(headerHex)

	s := NewSolver(1)
	s.SetHeader(header)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Solve(uint32(i*1000), 1000)
	}
}

func TestDifficultyCheck(t *testing.T) {
	// Test hash (all zeros should pass any target)
	hash := [32]byte{}

	// Test target (difficulty 1)
	target := make([]byte, 32)
	target[0] = 0xff
	target[1] = 0xff

	if !CheckDifficulty(hash, target) {
		t.Error("Zero hash should pass any target")
	}

	// Test harder target
	hash[31] = 0x01
	target[31] = 0x02

	if !CheckDifficulty(hash, target) {
		t.Error("Hash 0x01 should pass target 0x02")
	}

	// Test failing case
	hash[31] = 0x03
	target[31] = 0x02

	if CheckDifficulty(hash, target) {
		t.Error("Hash 0x03 should not pass target 0x02")
	}
}
