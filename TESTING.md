# Shock Testing Environment

This document describes the Docker Compose testing environment for the Shock repository, including setup instructions, available commands, and common testing scenarios.

## Overview

The Shock testing environment uses Docker Compose to create a consistent and isolated environment for running tests. It includes:

- A service for running Go tests
- A MongoDB instance for test data
- Volume mappings to persist test results and coverage reports
- Environment variables for test configuration
- Dedicated output directory for test results and coverage reports

## Setup Instructions

### Prerequisites

- Docker and Docker Compose installed on your system
- Git (to clone the repository)

### Initial Setup

1. Clone the Shock repository (if you haven't already):
   ```bash
   git clone https://github.com/MG-RAST/Shock.git
   cd Shock
   ```

2. Make the test script executable (if it's not already):
   ```bash
   chmod +x run-tests.sh
   ```

## Available Commands

The `run-tests.sh` script provides a simple interface for running tests:

```bash
./run-tests.sh [command] [options]
```

### Commands

- `all`: Run all tests (default)
- `package [path]`: Run tests for a specific package (e.g., `./shock-server/node`)
- `coverage`: Generate test coverage report
- `output-dir [path]`: Set custom directory for test outputs (default: `./test-output`)
- `clean`: Clean up test environment
- `help`: Show help message

### Options

- `-v, --verbose`: Enable verbose output
- `-h, --help`: Show help message

## Common Testing Scenarios

### Running All Tests

To run all tests in the repository:

```bash
./run-tests.sh all
```

### Running Tests for a Specific Package

To run tests for a specific package:

```bash
./run-tests.sh package ./shock-server/node
```

### Generating a Coverage Report

To generate a test coverage report:

```bash
./run-tests.sh coverage
```

This will create an HTML coverage report (`test-output/coverage.html`) in the project's test output directory and attempt to open it in your default browser. The test results will also be saved to `test-output/test-results.log`.

### Specifying a Custom Output Directory

To specify a custom directory for test outputs:

```bash
./run-tests.sh output-dir ./my-test-results
```

This will save all test outputs (results log and coverage reports) to the specified directory.

### Cleaning Up the Test Environment

To clean up the test environment (remove containers, volumes, and coverage files):

```bash
./run-tests.sh clean
```

## Test Environment Details

### Docker Compose Configuration

The test environment is defined in `docker-compose.test.yml` and includes:

- **shock-test**: Service for running Go tests
  - Built from `Dockerfile.test`
  - Mounts the project directory for live testing
  - Configurable via environment variables

- **shock-mongo-test**: MongoDB instance for testing
  - Uses MongoDB 3.6
  - Includes health checks to ensure it's ready before tests run

### Environment Variables

The test environment uses the following environment variables:

- `GO_TEST_ARGS`: Arguments to pass to `go test` (default: `./...`)
- `GO_TEST_COVERAGE`: Whether to generate coverage reports (default: `false`)
- `GO_TEST_COVERAGE_FILE`: Path to coverage output file (default: `/test-output/coverage.out`)
- `GO_TEST_COVERAGE_HTML`: Path to HTML coverage report (default: `/test-output/coverage.html`)
- `GO_TEST_RESULTS_FILE`: Path to test results log file (default: `/test-output/test-results.log`)

These can be set in the environment or passed via the `run-tests.sh` script.

## Extending the Test Environment

### Adding New Test Dependencies

If your tests require additional services (e.g., a mock API server), you can add them to the `docker-compose.test.yml` file:

```yaml
services:
  # Existing services...
  
  mock-api:
    image: mockserver/mockserver
    ports:
      - "1080"
    environment:
      - MOCKSERVER_INITIALIZATION_JSON_PATH=/config/mockserver.json
    volumes:
      - ./test/mockserver.json:/config/mockserver.json
```

### Customizing Test Configuration

The test configuration is defined in `Dockerfile.test` and can be customized as needed. For example, to add additional test dependencies:

```dockerfile
# Install additional dependencies
RUN go get -u github.com/some/package
```

## Troubleshooting

### Tests Fail to Connect to MongoDB

If tests fail with connection errors to MongoDB, ensure that:

1. The MongoDB service is running and healthy:
   ```bash
   docker-compose -f docker-compose.test.yml ps
   ```

2. The MongoDB connection string in the test configuration is correct:
   ```
   mongodb://shock-mongo-test:27017/shock_test
   ```

### Permission Issues with Test Files

If you encounter permission issues with test files:

1. Check the ownership of files created by Docker:
   ```bash
   ls -la
   ```

2. If necessary, adjust permissions:
   ```bash
   sudo chown -R $(id -u):$(id -g) .
   ```

### Coverage Report Not Generated

If the coverage report is not generated:

1. Ensure the test run completed successfully
2. Check for errors in the test output
3. Verify that the test output directory exists and has the correct permissions:
   ```bash
   ls -la test-output/
   ```

4. Verify that the coverage file was created:
   ```bash
   ls -la test-output/coverage.out
   ```

5. Manually generate the HTML report:
   ```bash
   go tool cover -html=test-output/coverage.out -o test-output/coverage.html
   ```

### Accessing Test Results

Test results are saved to a log file in the test output directory. To view the test results:

```bash
cat test-output/test-results.log
```

For a quick summary of test results:

```bash
grep -E "^(PASS|FAIL)" test-output/test-results.log | sort | uniq -c
```

## Best Practices for Writing Tests

1. **Use Mocks**: Use the provided mock implementations for database and file operations
2. **Isolate Tests**: Each test should be independent and not rely on the state from other tests
3. **Clean Up**: Always clean up resources created during tests
4. **Use Test Helpers**: Utilize the test utilities in `shock-server/test/util`
5. **Test Edge Cases**: Include tests for error conditions and edge cases