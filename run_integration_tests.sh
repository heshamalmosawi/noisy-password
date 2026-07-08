#!/bin/bash

# Run integration test 50 times and report failures

RUNS=500
PASSED=0
FAILED=0
FAILED_RUNS=()

echo "Running integration test $RUNS times..."
echo

for i in $(seq 1 $RUNS); do
    echo -n "Run $i/$RUNS: "
    if go test ./test  -v > /tmp/test_output_$i.log 2>&1; then
        echo "✓ PASSED"
        ((PASSED++))
    else
        echo "✗ FAILED"
        ((FAILED++))
        FAILED_RUNS+=($i)
    fi
done

echo
echo "================================"
echo "Results: $PASSED passed, $FAILED failed"
echo "================================"

if [ $FAILED -gt 0 ]; then
    echo
    echo "Failed runs: ${FAILED_RUNS[@]}"
    echo
    echo "First failed test output:"
    cat /tmp/test_output_${FAILED_RUNS[0]}.log
    exit 1
else
    echo "All runs passed! 🎉"
    exit 0
fi
