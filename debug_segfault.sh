#!/bin/bash

echo "=== Debugging Segmentation Fault ==="
echo

# Check which solver library was built
echo "1. Checking solver library:"
if [ -f solver/tromp/libcuckoo_lean.a ]; then
    echo "✓ libcuckoo_lean.a exists"
    ls -la solver/tromp/libcuckoo_lean.a
    echo
    echo "Object files in library:"
    ar -t solver/tromp/libcuckoo_lean.a
else
    echo "✗ libcuckoo_lean.a NOT FOUND"
fi

echo
echo "2. Checking which source was compiled:"
ls -la solver/tromp/*.o 2>/dev/null || echo "No object files found"

echo
echo "3. Running with gdb to catch segfault:"
echo "Commands to run:"
echo "  gdb ./bin/miner"
echo "  (gdb) run -u i5-6600 -t 1"
echo "  (gdb) bt"
echo "  (gdb) quit"

echo
echo "4. Quick test with strace:"
echo "  strace -e trace=memory ./bin/miner -u test -t 1 2>&1 | tail -50"

echo
echo "5. Check core dump (if enabled):"
echo "  ulimit -c unlimited"
echo "  ./bin/miner -u test -t 1"
echo "  gdb ./bin/miner core"
