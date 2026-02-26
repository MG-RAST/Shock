#!/bin/bash

# run-tests.sh - Script for running Shock tests in Docker environment
# 
# Usage:
#   ./run-tests.sh [command] [options]
#
# Commands:
#   all                 Run all tests (default)
#   package [path]      Run tests for a specific package (e.g., ./shock-server/node)
#   coverage            Generate test coverage report
#   clean               Clean up test environment
#   help                Show this help message
#
# Options:
#   -v, --verbose       Enable verbose output
#   -h, --help          Show this help message

set -e

# Default values
VERBOSE=false
COMMAND="all"
PACKAGE_PATH="./..."
COVERAGE=false
OUTPUT_DIR="./test-output"
COVERAGE_FILE="${OUTPUT_DIR}/coverage.out"
COVERAGE_HTML="${OUTPUT_DIR}/coverage.html"
TEST_RESULTS_FILE="${OUTPUT_DIR}/test-results.log"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        all)
            COMMAND="all"
            shift
            ;;
        package)
            COMMAND="package"
            if [[ $# -gt 1 && ! "$2" =~ ^- ]]; then
                PACKAGE_PATH="$2"
                shift 2
            else
                echo "Error: package command requires a path argument"
                exit 1
            fi
            ;;
        coverage)
            COMMAND="coverage"
            COVERAGE=true
            shift
            ;;
        output-dir)
            if [[ $# -gt 1 && ! "$2" =~ ^- ]]; then
                OUTPUT_DIR="$2"
                COVERAGE_FILE="${OUTPUT_DIR}/coverage.out"
                COVERAGE_HTML="${OUTPUT_DIR}/coverage.html"
                TEST_RESULTS_FILE="${OUTPUT_DIR}/test-results.log"
                shift 2
            else
                echo "Error: output-dir command requires a path argument"
                exit 1
            fi
            ;;
        clean)
            COMMAND="clean"
            shift
            ;;
        help)
            COMMAND="help"
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -h|--help)
            COMMAND="help"
            shift
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Show help message
show_help() {
    cat << EOF
Usage:
  ./run-tests.sh [command] [options]

Commands:
  all                 Run all tests (default)
  package [path]      Run tests for a specific package (e.g., ./shock-server/node)
  coverage            Generate test coverage report
  output-dir [path]   Set custom directory for test outputs (default: ./test-output)
  clean               Clean up test environment
  help                Show this help message

Options:
  -v, --verbose       Enable verbose output
  -h, --help          Show this help message

Test Output Files:
  Test results will be saved to: ${TEST_RESULTS_FILE}
  Coverage data will be saved to: ${COVERAGE_FILE}
  HTML coverage report will be saved to: ${COVERAGE_HTML}
EOF
}

# Build the test environment
build_environment() {
    echo "Building test environment..."
    docker-compose -f docker-compose.test.yml build
}

# Ensure output directory exists
ensure_output_dir() {
    mkdir -p "$OUTPUT_DIR"
    chmod 777 "$OUTPUT_DIR"
    echo "Test outputs will be saved to: $OUTPUT_DIR"
}

# Run tests
run_tests() {
    local test_args="$1"
    local coverage="$2"
    
    echo "Running tests: $test_args"
    
    # Ensure output directory exists
    ensure_output_dir
    
    # Set environment variables for docker-compose
    export GO_TEST_ARGS="$test_args"
    export GO_TEST_COVERAGE="$coverage"
    export GO_TEST_COVERAGE_FILE="$COVERAGE_FILE"
    export GO_TEST_COVERAGE_HTML="$COVERAGE_HTML"
    export GO_TEST_RESULTS_FILE="$TEST_RESULTS_FILE"
    
    # Run tests with docker-compose
    if [ "$VERBOSE" = true ]; then
        docker-compose -f docker-compose.test.yml up --abort-on-container-exit
    else
        docker-compose -f docker-compose.test.yml up --abort-on-container-exit > /dev/null 2>&1
        echo "Tests completed. Check the test results above."
    fi
}

# Clean up the test environment
clean_environment() {
    echo "Cleaning up test environment..."
    docker-compose -f docker-compose.test.yml down -v
    
    # Remove test output directory if it exists
    if [ -d "$OUTPUT_DIR" ]; then
        rm -rf "$OUTPUT_DIR"
        echo "Removed test output directory: $OUTPUT_DIR"
    fi
    
    echo "Test environment cleaned up."
}

# Main execution
case "$COMMAND" in
    all)
        build_environment
        run_tests "./..." "false"
        ;;
    package)
        build_environment
        run_tests "$PACKAGE_PATH" "false"
        ;;
    coverage)
        build_environment
        run_tests "./..." "true"
        
        # Check if test output files were generated
        if [ -f "$TEST_RESULTS_FILE" ]; then
            echo "Test results saved to: $TEST_RESULTS_FILE"
        else
            echo "Warning: Test results file was not generated."
        fi
        
        if [ -f "$COVERAGE_HTML" ]; then
            echo "Coverage report generated: $COVERAGE_HTML"
            
            # Open the coverage report if a browser is available
            if command -v xdg-open > /dev/null 2>&1; then
                xdg-open "$COVERAGE_HTML"
            elif command -v open > /dev/null 2>&1; then
                open "$COVERAGE_HTML"
            else
                echo "To view the coverage report, open $COVERAGE_HTML in your browser."
            fi
        else
            echo "Error: Coverage report was not generated."
            echo "This could be because the tests didn't run correctly or the file wasn't properly mounted from the container."
            echo "Try running with the --verbose flag to see more details."
        fi
        
        # Display summary of test results if available
        if [ -f "$TEST_RESULTS_FILE" ]; then
            echo "=== Test Summary ==="
            grep -E "^(PASS|FAIL)" "$TEST_RESULTS_FILE" | sort | uniq -c
            echo "=================="
        fi
        ;;
    clean)
        clean_environment
        ;;
    help)
        show_help
        ;;
esac

exit 0