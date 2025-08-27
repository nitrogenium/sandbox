#!/bin/bash

# Force build with x86 lean solver

set -e

echo "=== Force Building x86 Lean Solver ==="
echo

cd solver/tromp

echo "1. Cleaning old files..."
make clean 2>/dev/null || true
rm -f *.o *.a

echo
echo "2. Compiling cuckoo_lean.cpp..."
g++ -O3 -march=native -mavx2 -msse4.2 -std=c++14 -Wall -pthread \
    -I. -Icuckoo-orig/src -Icuckoo-orig/src/crypto \
    -c cuckoo_lean.cpp -o cuckoo_lean.o

echo
echo "3. Creating library..."
ar rcs libcuckoo_lean.a cuckoo_lean.o

echo
echo "4. Verifying library..."
nm libcuckoo_lean.a | grep -q "cuckoo_ctx" && echo "✓ Found cuckoo_ctx (lean solver)" || echo "✗ Missing cuckoo_ctx"
nm libcuckoo_lean.a | grep -q "cuckoo_solve" && echo "✓ Found cuckoo_solve" || echo "✗ Missing cuckoo_solve"

cd ../..

echo
echo "5. Rebuilding Go miner..."
go clean -cache
go build -o bin/miner cmd/miner/main.go

echo
echo "=== Build Complete ==="
echo "Now try: ./bin/miner -u test -t 1"
