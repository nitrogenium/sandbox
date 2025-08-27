# Quick Start Guide

## 1. Build the Miner

```bash
cd go-rebuild
./build.sh
```

## 2. Run Basic Mining

```bash
# Simple run with default settings
./bin/miner -pool tht.mine-n-krush.org:8333 -user YOUR_WALLET.worker1

# With specific thread count
./bin/miner -pool tht.mine-n-krush.org:8333 -user YOUR_WALLET.worker1 -threads 16
```

## 3. Optimize for AMD Ryzen 7950X

```bash
# Run optimization script (requires sudo)
sudo ./optimize_amd.sh

# Run optimized miner on single CCD (better efficiency)
taskset -c 0-15 ./bin/miner -pool tht.mine-n-krush.org:8333 -user YOUR_WALLET.worker1 -threads 16

# Or use all cores (maximum hashrate)
./bin/miner -pool tht.mine-n-krush.org:8333 -user YOUR_WALLET.worker1 -threads 32
```

## 4. Monitor Performance

Watch miner output for:
- `cycles/s`: Graph traversal rate (higher is better)
- `solutions/s`: Valid cycles found (should be > 0)
- `sharesAccepted`: Successfully submitted shares

## Common Issues

### Build fails with C++ errors
```bash
# Install build dependencies
sudo apt-get install build-essential g++ make
```

### Low performance
1. Check CPU governor: `cat /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor`
2. Should show "performance" not "powersave"
3. Run optimization script with sudo

### Connection errors
- Check pool address and port
- Verify username format (usually WALLET.WORKER)
- Check firewall settings
