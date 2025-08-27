#!/bin/bash

echo "=== Running Solver Tests ==="
echo

cd solver/tromp

# Test 1: Direct Tromp's code
echo "1. Testing Tromp's solver directly..."
make -f test_makefile clean
make -f test_makefile
if [ -f test_tromp ]; then
    ./test_tromp
    echo
else
    echo "✗ Failed to build test_tromp"
fi

# Test 2: C API wrapper
echo "2. Testing C API wrapper..."
make -f test_capi_makefile clean  
make -f test_capi_makefile
if [ -f test_capi ]; then
    ./test_capi
    echo
else
    echo "✗ Failed to build test_capi"
fi

# Test 3: Check current library
echo "3. Checking current library..."
if [ -f libcuckoo_lean.a ]; then
    echo "Library size: $(ls -lh libcuckoo_lean.a | awk '{print $5}')"
    echo "Symbols:"
    nm libcuckoo_lean.a | grep -E "(cuckoo_solve|cuckoo_init|cuckoo_verify)" | head -10
else
    echo "✗ No library found"
fi

cd ../..
echo
echo "=== Tests Complete ==="
echo
echo "If tests pass here but the miner still segfaults, the issue is likely in Go/CGO integration."
