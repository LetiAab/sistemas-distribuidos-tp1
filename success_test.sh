#!/bin/bash

# Simple integration test - compares sequential vs distributed results

set -e

echo "Running integration test: compare sequential vs distributed results"

cd src/

# Clean up any previous results
echo "Cleaning up previous results..."
rm -f mr-out-* files/mr-out-* sequential.txt distributed.txt 2>/dev/null || true

# Run sequential
echo "Running sequential version..."
go run secuencial.go ../books/pg-*.txt
echo "Sorting sequential results..."
sort mr-out-0 > sequential.txt

# Build plugin if needed
if [ ! -f plugins/wc.so ]; then
    echo "Building wc plugin..."
    echo "$ go build -buildmode=plugin -o plugins/wc.so plugins/wc.go"
    go build -buildmode=plugin -o plugins/wc.so plugins/wc.go
fi

# Run distributed
echo "Running distributed version..."
go run coordinator/coordinator.go ../books/pg-*.txt > /dev/null 2>&1 &
COORD_PID=$!
echo "Coordinator started (PID: $COORD_PID)"
sleep 2

echo "Starting 3 workers..."
go run worker/worker.go plugins/wc.so > /dev/null 2>&1 &
go run worker/worker.go plugins/wc.so > /dev/null 2>&1 &
go run worker/worker.go plugins/wc.so > /dev/null 2>&1 &

# Wait for completion (simple approach)
echo "Waiting for processing to complete (30s)..."
sleep 30

# Kill processes
echo "Stopping processes..."
kill $COORD_PID 2>/dev/null || true
pkill -f worker.go 2>/dev/null || true

# Compare results
echo "Concatenating and sorting distributed results..."
cat files/mr-out-* | sort > distributed.txt

echo "Comparing results..."
if diff sequential.txt distributed.txt > /dev/null; then
    echo "✅ PASSED: Results match"
else
    echo "❌ FAILED: Results differ"
    exit 1
fi