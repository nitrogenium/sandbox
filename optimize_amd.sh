#!/bin/bash

# Performance optimization script for AMD Ryzen 7950X

echo "=== AMD Ryzen 7950X Optimization Script ==="

# Check if running as root for some optimizations
if [ "$EUID" -ne 0 ]; then 
    echo "Note: Some optimizations require root access. Run with sudo for full optimization."
fi

# 1. Set CPU governor to performance
echo "Setting CPU governor to performance..."
for cpu in /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor; do
    echo "performance" | sudo tee $cpu > /dev/null 2>&1
done

# 2. Disable CPU frequency scaling
echo "Disabling frequency scaling..."
sudo cpupower frequency-set -g performance 2>/dev/null

# 3. Enable huge pages (2MB pages)
echo "Configuring huge pages..."
sudo sysctl -w vm.nr_hugepages=2048 2>/dev/null
sudo sysctl -w vm.hugetlb_shm_group=0 2>/dev/null

# 4. Set memory allocator to use huge pages
export MALLOC_ARENA_MAX=2
export MALLOC_MMAP_THRESHOLD_=131072
export MALLOC_TRIM_THRESHOLD_=131072
export MALLOC_TOP_PAD_=131072
export MALLOC_MMAP_MAX_=65536
export MALLOC_CHECK_=0

# 5. NUMA optimization for dual CCD
echo "Checking NUMA configuration..."
numactl --hardware 2>/dev/null

# 6. Disable SMT for better per-core performance (optional)
# echo "Disabling SMT (Hyperthreading)..."
# echo off | sudo tee /sys/devices/system/cpu/smt/control 2>/dev/null

# 7. Set process priority
echo "Setting process scheduling..."
export GOMAXPROCS=16  # Use physical cores only

# 8. Cache optimization
echo 3 | sudo tee /proc/sys/vm/drop_caches 2>/dev/null

# Build optimized binary
echo "Building optimized miner..."
export CGO_CFLAGS="-O3 -march=znver4 -mtune=znver4 -msse4.2 -mavx2"
export CGO_CXXFLAGS="-O3 -march=znver4 -mtune=znver4 -msse4.2 -mavx2 -std=c++14"
export CGO_LDFLAGS="-lpthread"

cd "$(dirname "$0")"
./build.sh

echo ""
echo "=== Optimization Complete ==="
echo ""
echo "Run miner with optimal settings:"
echo "  Single CCD (better cache locality):"
echo "    taskset -c 0-15 ./bin/miner -pool <host:port> -user <user> -threads 16"
echo ""
echo "  Dual CCD (maximum throughput):"
echo "    numactl --cpunodebind=0,1 ./bin/miner -pool <host:port> -user <user> -threads 32"
echo ""
echo "Monitor performance with:"
echo "  watch -n 1 'sensors | grep -A 4 k10temp'"
echo "  htop -d 10"
