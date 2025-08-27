# Go Cuckoo Cycle Miner

High-performance Cuckoo Cycle miner implementation in Go with C++ solver backend.

## Features

- **Optimized Cuckoo Solver**: Uses John Tromp's lean solver via cgo
- **Stratum Protocol**: Full Stratum mining protocol support
- **Multi-threading**: Efficient parallel mining with goroutines
- **Auto-reconnect**: Automatic pool reconnection on failures
- **Real-time Stats**: Mining performance monitoring

## Architecture

```
┌─────────────────┐     ┌──────────────────┐
│   Main (Go)     │────▶│ Stratum Client   │────▶ Mining Pool
│                 │     │     (Go)         │      (TCP/JSON-RPC)
└────────┬────────┘     └──────────────────┘
         │
         ▼
┌─────────────────┐     ┌──────────────────┐
│ Mining Workers  │────▶│  Solver (cgo)    │
│   (Goroutines)  │     │                  │
└─────────────────┘     └────────┬─────────┘
                                 │
                                 ▼
                        ┌──────────────────┐
                        │ Tromp's Lean C++ │
                        │  Cuckoo Solver   │
                        └──────────────────┘
```

## Building

### Prerequisites

- Go 1.22 or later
- GCC/G++ with C++14 support
- Make

### Build Steps

```bash
# Clone and build
./build.sh
```

## Usage

```bash
./bin/miner -pool <host:port> -user <username> [-pass <password>] [-threads <n>]
```

### Example

```bash
# Connect to pool with 16 threads
./bin/miner -pool tht.mine-n-krush.org:8333 -user myworker -threads 16
```

## Configuration

- `-pool`: Stratum pool address (required)
- `-user`: Worker username (required)  
- `-pass`: Worker password (default: "x")
- `-threads`: Number of mining threads (default: CPU cores)
- `-debug`: Enable debug logging

## Performance Optimization

### For AMD Ryzen 7950X:

1. **CPU Affinity**: Pin threads to specific cores
2. **NUMA Awareness**: Separate threads between CCDs
3. **Memory**: Use huge pages for better TLB performance

```bash
# Enable huge pages
sudo sysctl -w vm.nr_hugepages=2048

# Run with taskset for CPU affinity
taskset -c 0-15 ./bin/miner -pool ... -threads 16
```

### Compiler Optimizations

The solver is compiled with:
- `-O3`: Maximum optimization
- `-march=native`: CPU-specific instructions
- `-pthread`: Threading support

## Monitoring

The miner prints statistics every 10 seconds:
- Cycles/second: Graph traversal rate
- Solutions/second: Valid cycle finding rate
- Shares accepted/rejected: Pool submission stats

## Development

### Project Structure

```
go-rebuild/
├── cmd/miner/         # Main miner executable
├── pkg/
│   ├── solver/        # Go wrapper for C++ solver
│   └── stratum/       # Stratum protocol implementation
├── solver/tromp/      # C++ Cuckoo solver
└── build.sh          # Build script
```

### Testing

```bash
# Run tests
go test ./...

# Benchmark solver
go test -bench=. ./pkg/solver
```

## License

Based on John Tromp's Cuckoo Cycle implementation.
See LICENSE for details.
