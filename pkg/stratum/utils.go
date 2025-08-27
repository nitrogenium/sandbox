package stratum

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
)

// BuildHeader constructs 80-byte header from Stratum work
func BuildHeader(work *Work, extraNonce2 string) ([]byte, error) {
	// Decode hex strings
	version, err := hex.DecodeString(work.Version)
	if err != nil {
		return nil, fmt.Errorf("invalid version: %w", err)
	}

	prevHash, err := hex.DecodeString(work.PrevHash)
	if err != nil {
		return nil, fmt.Errorf("invalid prevhash: %w", err)
	}

	// Build coinbase transaction
	coinbase := work.Coinbase1 + work.ExtraNonce1 + extraNonce2 + work.Coinbase2
	coinbaseBytes, err := hex.DecodeString(coinbase)
	if err != nil {
		return nil, fmt.Errorf("invalid coinbase: %w", err)
	}

	// Calculate coinbase hash
	coinbaseHash := sha256d(coinbaseBytes)

	// Build merkle root
	merkleRoot := coinbaseHash
	for _, branch := range work.MerkleBranch {
		branchBytes, err := hex.DecodeString(branch)
		if err != nil {
			return nil, fmt.Errorf("invalid merkle branch: %w", err)
		}
		merkleRoot = sha256d(append(merkleRoot, branchBytes...))
	}

	ntime, err := hex.DecodeString(work.NTime)
	if err != nil {
		return nil, fmt.Errorf("invalid ntime: %w", err)
	}

	nbits, err := hex.DecodeString(work.NBits)
	if err != nil {
		return nil, fmt.Errorf("invalid nbits: %w", err)
	}

	// Build 80-byte header
	header := make([]byte, 80)
	copy(header[0:4], reverseBytes(version))   // Version
	copy(header[4:36], reverseBytes(prevHash)) // Previous block hash
	copy(header[36:68], merkleRoot)            // Merkle root
	copy(header[68:72], reverseBytes(ntime))   // Timestamp
	copy(header[72:76], reverseBytes(nbits))   // Bits
	// Nonce (76:80) will be filled by miner

	return header, nil
}

// DifficultyToTarget converts pool difficulty to 32-byte target
func DifficultyToTarget(difficulty float64) []byte {
	// Pool difficulty 1 = 0x00000000ffff0000000000000000000000000000000000000000000000000000
	// Target = bdiff1 / difficulty

	bdiff1 := new(big.Int)
	bdiff1.SetString("00000000ffff0000000000000000000000000000000000000000000000000000", 16)

	// Convert difficulty to big.Int with precision
	diffBig := new(big.Float).SetFloat64(difficulty)
	diffInt := new(big.Int)
	diffBig.Int(diffInt)

	if diffInt.Sign() <= 0 {
		diffInt.SetInt64(1)
	}

	target := new(big.Int).Div(bdiff1, diffInt)

	// Ensure 32 bytes
	targetBytes := make([]byte, 32)
	targetBigBytes := target.Bytes()
	copy(targetBytes[32-len(targetBigBytes):], targetBigBytes)

	return targetBytes
}

// CompactToTarget converts nBits compact format to 32-byte target
func CompactToTarget(compact uint32) []byte {
	// Extract mantissa and exponent
	mantissa := compact & 0x007fffff
	exponent := compact >> 24

	// Calculate target
	target := new(big.Int).SetUint64(uint64(mantissa))
	target.Lsh(target, uint(8*(exponent-3)))

	// Ensure 32 bytes
	targetBytes := make([]byte, 32)
	targetBigBytes := target.Bytes()
	copy(targetBytes[32-len(targetBigBytes):], targetBigBytes)

	return targetBytes
}

// CheckTarget verifies if hash meets target difficulty
func CheckTarget(hash, target []byte) bool {
	// Compare as unsigned 256-bit integers (big-endian)
	for i := 0; i < 32; i++ {
		if hash[i] < target[i] {
			return true
		}
		if hash[i] > target[i] {
			return false
		}
	}
	return true
}

// GenerateExtraNonce2 generates extraNonce2 of required size
func GenerateExtraNonce2(size int, counter uint64) string {
	bytes := make([]byte, size)
	
	// Write counter bytes, handling cases where size < 8
	if size >= 8 {
		binary.LittleEndian.PutUint64(bytes, counter)
	} else {
		// For smaller sizes, write only lower bytes
		counterBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(counterBytes, counter)
		copy(bytes, counterBytes[:size])
	}
	
	return hex.EncodeToString(bytes)
}

// sha256d performs double SHA256
func sha256d(data []byte) []byte {
	h1 := sha256.Sum256(data)
	h2 := sha256.Sum256(h1[:])
	return h2[:]
}

// reverseBytes reverses byte slice
func reverseBytes(data []byte) []byte {
	result := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		result[i] = data[len(data)-1-i]
	}
	return result
}

// UpdateNTime increments nTime by seconds
func UpdateNTime(ntime string, seconds int) (string, error) {
	bytes, err := hex.DecodeString(ntime)
	if err != nil {
		return "", err
	}

	// Convert to uint32 (little-endian)
	timestamp := binary.LittleEndian.Uint32(reverseBytes(bytes))
	timestamp += uint32(seconds)

	// Convert back
	newBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(newBytes, timestamp)

	return hex.EncodeToString(reverseBytes(newBytes)), nil
}
