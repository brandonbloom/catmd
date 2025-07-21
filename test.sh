#!/bin/bash

set -e

echo "Building catmd..."
go build -o catmd .

echo "Running tests..."

# Test directory
TEST_DIR="test"
mkdir -p "$TEST_DIR"

# Function to run a test
run_test() {
    local test_name="$1"
    local test_dir="$TEST_DIR/$test_name"
    
    echo "Running test: $test_name"
    
    # Check for index.md or a.md as starting file
    local start_file=""
    if [ -f "$test_dir/input/index.md" ]; then
        start_file="$test_dir/input/index.md"
    elif [ -f "$test_dir/input/a.md" ]; then
        start_file="$test_dir/input/a.md"
    fi
    
    if [ -n "$start_file" ]; then
        # Check if test has custom options
        if [ "$test_name" = "scope-option" ]; then
            ./catmd --scope "$test_dir/input/docs" "$start_file" > "$test_dir/actual.md"
        elif [ "$test_name" = "output-option" ]; then
            ./catmd -o "$test_dir/actual.md" "$start_file"
        else
            ./catmd "$start_file" > "$test_dir/actual.md"
        fi
        
        if diff -q "$test_dir/expected.md" "$test_dir/actual.md" > /dev/null; then
            echo "  ✓ PASSED"
        else
            echo "  ✗ FAILED"
            echo "    Differences:"
            diff "$test_dir/expected.md" "$test_dir/actual.md" || true
        fi
    else
        echo "  ✗ SKIPPED (missing input file)"
    fi
}

# Run all tests
for test_dir in "$TEST_DIR"/*; do
    if [ -d "$test_dir" ]; then
        test_name=$(basename "$test_dir")
        run_test "$test_name"
    fi
done

echo "Tests completed."
