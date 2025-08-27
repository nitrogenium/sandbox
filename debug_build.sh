#!/bin/bash

echo "=== Debug Build for Segfault Investigation ==="

# Go to solver directory
cd solver/tromp

# Clean old files
echo "1. Cleaning..."
make clean || true
rm -f *.o *.a

# Build with debug symbols and no optimization
echo "2. Building with debug symbols..."
g++ -g -O0 -march=native -std=c++14 -Wall -pthread \
    -I. -Icuckoo-orig/src -Icuckoo-orig/src/crypto \
    -c cuckoo_lean.cpp -o cuckoo_lean.o

if [ $? -ne 0 ]; then
    echo "✗ Compilation failed"
    exit 1
fi

# Create library
ar rcs libcuckoo_lean.a cuckoo_lean.o

# Back to root
cd ../..

# Build Go with debug
echo "3. Building Go miner with debug..."
go build -gcflags="all=-N -l" -o bin/miner-debug cmd/miner/main.go

if [ -f bin/miner-debug ]; then
    echo "✓ Debug build successful"
    echo
    echo "Now run with gdb:"
    echo "  gdb ./bin/miner-debug"
    echo "  (gdb) run -u test -t 1"
    echo "  (gdb) bt"
else
    echo "✗ Build failed"
    exit 1
fi
