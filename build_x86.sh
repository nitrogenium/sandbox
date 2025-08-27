#!/bin/bash

# Build script for x86_64 Linux systems
# Optimized for Intel i5-6600 (Skylake)

set -e

echo "=== Building Go Cuckoo Miner for x86_64 ==="
echo

# Check architecture
ARCH=$(uname -m)
if [ "$ARCH" != "x86_64" ]; then
    echo "Error: This script is for x86_64 architecture only"
    echo "Detected: $ARCH"
    exit 1
fi

# Check for required tools
echo "Checking dependencies..."
for cmd in gcc g++ go make; do
    if ! command -v $cmd &> /dev/null; then
        echo "Error: $cmd is not installed"
        exit 1
    fi
done

echo "✓ All dependencies found"
echo

# Detect CPU features
echo "Detecting CPU features..."
CPU_FLAGS=""
if grep -q avx2 /proc/cpuinfo; then
    CPU_FLAGS="$CPU_FLAGS -mavx2"
    echo "✓ AVX2 supported"
fi
if grep -q sse4_2 /proc/cpuinfo; then
    CPU_FLAGS="$CPU_FLAGS -msse4.2"
    echo "✓ SSE4.2 supported"
fi

# Set optimization flags for Intel i5-6600 (Skylake)
export CC=gcc
export CXX=g++
export CGO_ENABLED=1
export CGO_CFLAGS="-O3 -march=skylake -mtune=skylake $CPU_FLAGS"
export CGO_CXXFLAGS="-O3 -march=skylake -mtune=skylake $CPU_FLAGS -std=c++14"
export CGO_LDFLAGS="-lpthread"

echo
echo "Building C++ Cuckoo solver..."
cd solver/tromp

# Clean previous build
make clean 2>/dev/null || true

# Build with x86 optimizations
make ARCH_FLAGS="-march=skylake -mtune=skylake $CPU_FLAGS" CPPFLAGS=""

if [ ! -f libcuckoo_lean.a ]; then
    echo "Error: Failed to build C++ solver"
    exit 1
fi

echo "✓ C++ solver built successfully"
cd ../..

echo
echo "Building Go miner..."

# Download dependencies
go mod download

# Build miner
go build -ldflags="-s -w" -o bin/miner cmd/miner/main.go

if [ ! -f bin/miner ]; then
    echo "Error: Failed to build miner"
    exit 1
fi

# Make executable
chmod +x bin/miner

echo "✓ Miner built successfully"
echo

# Create optimized start script
cat > start_x86.sh << 'EOF'
#!/bin/bash

# Optimizations for Intel i5-6600
export GOMAXPROCS=4
export MALLOC_ARENA_MAX=2

# Check CPU governor
GOVERNOR=$(cat /sys/devices/system/cpu/cpu0/cpufreq/scaling_governor 2>/dev/null || echo "unknown")
if [ "$GOVERNOR" != "performance" ]; then
    echo "Warning: CPU governor is '$GOVERNOR', recommend 'performance' for best results"
    echo "Run: sudo cpupower frequency-set -g performance"
fi

# Run miner
exec ./bin/miner "$@"
EOF

chmod +x start_x86.sh

echo "=== Build Complete ==="
echo
echo "To run the miner:"
echo "  ./start_x86.sh"
echo
echo "Options:"
echo "  -u <worker>  : Worker name (default: CPU-666)"
echo "  -t <threads> : Number of threads (default: 4)"
echo "  -debug       : Enable debug output"
echo
echo "Example:"
echo "  ./start_x86.sh -u i5-6600 -t 4"
echo
