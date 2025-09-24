#!/bin/bash

# Fault tolerance test - kills workers during execution to test recovery

set -e

echo "Running fault tolerance test: testing worker failure recovery"

cd src/

# Clean up any previous results
echo "Cleaning up previous results..."
rm -f mr-out-* files/mr-out-* sequential.txt distributed.txt 2>/dev/null || true

# Run sequential for comparison
echo "Running sequential version for comparison..."
go run secuencial.go ../books/pg-*.txt
echo "Sorting sequential results..."
sort mr-out-0 > sequential.txt

# Build plugin if needed
if [ ! -f plugins/wc.so ]; then
    echo "Building wc plugin..."
    go build -buildmode=plugin -o plugins/wc.so plugins/wc.go
fi

# Run distributed with fault injection
echo "Running distributed version with fault tolerance test..."
go run coordinator/coordinator.go ../books/pg-*.txt > /dev/null 2>&1 &
COORD_PID=$!
echo "Coordinator started (PID: $COORD_PID)"
sleep 2

echo "Starting 4 workers..."
go run worker/worker.go plugins/wc.so > /dev/null 2>&1 &
WORKER1_PID=$!
go run worker/worker.go plugins/wc.so > /dev/null 2>&1 &
WORKER2_PID=$!
go run worker/worker.go plugins/wc.so > /dev/null 2>&1 &
WORKER3_PID=$!
go run worker/worker.go plugins/wc.so > /dev/null 2>&1 &
WORKER4_PID=$!

echo "Workers started (PIDs: $WORKER1_PID $WORKER2_PID $WORKER3_PID $WORKER4_PID)"

# Let them work for a bit
echo "Letting workers process for 10 seconds..."
sleep 10

# Kill some workers to simulate failures
echo "Killing worker $WORKER1_PID"
kill $WORKER1_PID 2>/dev/null || true
sleep 3

echo "Killing worker $WORKER3_PID"  
kill $WORKER3_PID 2>/dev/null || true

# Wait for completion
echo "Waiting for remaining workers to complete (30 more seconds)..."
sleep 30

# Kill all remaining processes
echo "Stopping all processes..."
kill $COORD_PID 2>/dev/null || true
pkill -f worker.go 2>/dev/null || true

# Compare results
echo "Concatenating and sorting distributed results..."
if ls files/mr-out-* 1> /dev/null 2>&1; then
    cat files/mr-out-* | sort > distributed.txt
    
    echo "Comparing results after fault tolerance test..."
    if diff sequential.txt distributed.txt > /dev/null; then
        echo "✅ PASSED: System recovered and results match"
    else
        echo "❌ FAILED: Results differ after worker failures"
        exit 1
    fi
else
    echo "❌ FAILED: No output files generated"
    exit 1
fi