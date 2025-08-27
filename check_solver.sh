#!/bin/bash

echo "=== Checking Cuckoo Solver Build ==="
echo

echo "1. Architecture:"
uname -m
echo

echo "2. Object files in solver/tromp:"
ls -la solver/tromp/*.o 2>/dev/null || echo "No .o files found"
echo

echo "3. Contents of libcuckoo_lean.a:"
if [ -f solver/tromp/libcuckoo_lean.a ]; then
    ar -t solver/tromp/libcuckoo_lean.a
else
    echo "Library not found!"
fi
echo

echo "4. Symbols in library:"
if [ -f solver/tromp/libcuckoo_lean.a ]; then
    nm solver/tromp/libcuckoo_lean.a | grep "cuckoo_ctx" | head -10
fi
echo

echo "5. Makefile SOURCES detection:"
cd solver/tromp
make -n clean all 2>&1 | grep -E "(SOURCES|\.cpp)" | head -5
cd ../..
echo

echo "=== RECOMMENDATION ==="
echo "If library is missing or corrupted:"
echo "1. cd solver/tromp"
echo "2. make clean"
echo "3. make"
echo "4. cd ../.."
echo "5. go build -o bin/miner cmd/miner/main.go"
