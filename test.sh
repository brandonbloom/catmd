#!/bin/bash

set -e

echo "Building catmd..."
go build -o bin/catmd .
export PATH="$(pwd)/bin:$PATH"

echo "Running tests..."

# Test directory
TEST_DIR="test"
mkdir -p "$TEST_DIR"

# Function to run a test
run_test() {
    local test_name="$1"
    local test_dir="$TEST_DIR/$test_name"
    
    echo "Running test: $test_name"
    
    if [ -f "$test_dir/test.config" ]; then
        # Read config file (contains catmd arguments)
        local config_args=$(cat "$test_dir/test.config")
        # Check if args contain output flag
        if [[ "$config_args" == *"-o "* ]]; then
            (cd "$test_dir" && catmd $config_args)
        else
            (cd "$test_dir" && catmd $config_args > actual.md)
        fi
    elif [ -f "$test_dir/index.md" ]; then
        # Default: use index.md in test directory
        catmd "$test_dir/index.md" > "$test_dir/actual.md"
    else
        echo "  ✗ SKIPPED (no test.config and no index.md)"
        return
    fi
        
    if diff -q "$test_dir/expected.md" "$test_dir/actual.md" > /dev/null; then
        echo "  ✓ PASSED"
    else
        echo "  ✗ FAILED"
        echo "    Differences:"
        diff "$test_dir/expected.md" "$test_dir/actual.md" || true
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
