#!/bin/bash

# Complete clean and rebuild script

echo "=== Complete Clean and Rebuild ==="
echo

echo "1. Cleaning all build artifacts..."
rm -f bin/miner
cd solver/tromp
make clean 2>/dev/null || true
rm -f *.o *.a
cd ../..

echo "2. Cleaning Go cache..."
go clean -cache -modcache

echo "3. Rebuilding everything..."
./build_x86.sh

echo
echo "=== Done ==="
echo "Run with: ./bin/miner -u <worker> -t <threads>"
