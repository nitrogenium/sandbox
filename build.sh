#!/bin/bash

# Build script for Go Cuckoo miner

set -e

echo "Building Cuckoo solver library..."
cd solver/tromp
make clean
make
cd ../..

echo "Downloading Go dependencies..."
go mod tidy
go mod download

echo "Building miner..."
go build -o bin/miner cmd/miner/main.go

echo "Build complete! Run with: ./bin/miner -pool <host:port> -user <username>"
