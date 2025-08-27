#!/bin/bash

echo "=== Building Go Cuckoo Miner for x86_64 (FIXED) ==="
echo

# Detect CPU features
echo "Detecting CPU features..."
CPU_FLAGS=""
if grep -q avx2 /proc/cpuinfo 2>/dev/null; then
    CPU_FLAGS="$CPU_FLAGS -mavx2"
    echo "✓ AVX2 supported"
fi
if grep -q sse4_2 /proc/cpuinfo 2>/dev/null; then
    CPU_FLAGS="$CPU_FLAGS -msse4.2"
    echo "✓ SSE4.2 supported"
fi

# Set build environment
export CC=gcc
export CXX=g++
export CGO_ENABLED=1
export CGO_CFLAGS="-O3 -march=native $CPU_FLAGS"
export CGO_CXXFLAGS="-O3 -march=native $CPU_FLAGS -std=c++14"
export CGO_LDFLAGS="-lpthread"

# Build C++ solver
echo
echo "Building C++ Cuckoo solver..."
cd solver/tromp

# Always clean first
make clean 2>/dev/null || true
rm -f *.o *.a

# Build with explicit settings
echo "Compiling cuckoo_lean.cpp..."
$CXX $CGO_CXXFLAGS -pthread -I. -Icuckoo-orig/src -Icuckoo-orig/src/crypto \
     -c cuckoo_lean.cpp -o cuckoo_lean.o

if [ ! -f cuckoo_lean.o ]; then
    echo "Error: Failed to compile cuckoo_lean.cpp"
    exit 1
fi

echo "Creating static library..."
ar rcs libcuckoo_lean.a cuckoo_lean.o

if [ ! -f libcuckoo_lean.a ]; then
    echo "Error: Failed to create library"
    exit 1
fi

echo "✓ C++ solver built successfully"
cd ../..

# Build Go miner
echo
echo "Building Go miner..."
go build -o bin/miner cmd/miner/main.go

if [ -f bin/miner ]; then
    echo
    echo "✓ Build successful!"
    echo
    echo "Miner binary: bin/miner"
    echo
    echo "Run with:"
    echo "  ./bin/miner -u <worker> -t <threads>"
    echo
    echo "Example:"
    echo "  ./bin/miner -u myworker -t 4"
else
    echo
    echo "✗ Build failed"
    exit 1
fi
