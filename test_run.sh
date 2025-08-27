#!/bin/bash

echo "Starting test miner run..."

# Run miner for 10 seconds then kill it
./bin/miner -pool localhost:3333 -user testuser -threads 2 -debug &
MINER_PID=$!

echo "Miner PID: $MINER_PID"
echo "Running for 10 seconds..."

sleep 10

echo "Stopping miner..."
kill $MINER_PID 2>/dev/null

echo "Test complete"
