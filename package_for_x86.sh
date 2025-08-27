#!/bin/bash

# Script to package the miner for x86_64 deployment

PACKAGE_NAME="go-cuckoo-miner-x86.tar.gz"

echo "Creating deployment package for x86_64..."

# Create temporary directory
TEMP_DIR=$(mktemp -d)
PACKAGE_DIR="$TEMP_DIR/go-cuckoo-miner"

# Copy necessary files
mkdir -p "$PACKAGE_DIR"
cp -r cmd pkg solver go.mod go.sum build_x86.sh BUILD_X86.md QUICKSTART_X86.md "$PACKAGE_DIR/" 2>/dev/null || true

# Remove build artifacts
rm -f "$PACKAGE_DIR/solver/tromp/libcuckoo_lean.a"

# Clean build artifacts
find "$PACKAGE_DIR" -name "*.o" -delete
find "$PACKAGE_DIR" -name "*.a" -delete

# Create package
cd "$TEMP_DIR"
tar -czf "$PACKAGE_NAME" go-cuckoo-miner/

# Move to original directory
mv "$PACKAGE_NAME" "$OLDPWD/"
cd "$OLDPWD"

# Cleanup
rm -rf "$TEMP_DIR"

echo "✓ Package created: $PACKAGE_NAME"
echo
echo "To deploy on x86_64 Linux:"
echo "1. Copy $PACKAGE_NAME to target machine"
echo "2. tar -xzf $PACKAGE_NAME"
echo "3. cd go-cuckoo-miner"
echo "4. ./build_x86.sh"
echo "5. ./start_x86.sh"
