package logger_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/stretchr/testify/assert"
)

// setupFileLogger initializes the logger with file output to a temp dir
// and returns the temp dir path for cleanup.
func setupFileLogger(t *testing.T) string {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "logger-test")
	if err != nil {
		t.Fatal(err)
	}
	conf.LOG_OUTPUT = "file"
	conf.LOG_ROTATE = false
	conf.PATH_LOGS = tempDir
	logger.Initialize()
	return tempDir
}

// readLogFile reads and returns the contents of a log file, with a short
// delay to allow the async logger goroutine to process.
func readLogFile(t *testing.T, dir, name string) string {
	t.Helper()
	time.Sleep(300 * time.Millisecond)
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return ""
	}
	return string(data)
}

// TestInitialize tests the Initialize function with various LOG_OUTPUT modes
func TestInitialize(t *testing.T) {
	// Test with LOG_OUTPUT = "console"
	conf.LOG_OUTPUT = "console"
	conf.LOG_ROTATE = false
	conf.PATH_LOGS = ""
	logger.Initialize()
	assert.NotNil(t, logger.Log, "Logger should be initialized")

	// Test with LOG_OUTPUT = "file"
	tempDir, err := os.MkdirTemp("", "logger-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	conf.LOG_OUTPUT = "file"
	conf.PATH_LOGS = tempDir
	logger.Initialize()
	assert.NotNil(t, logger.Log, "Logger should be initialized")

	// Test with LOG_OUTPUT = "both"
	conf.LOG_OUTPUT = "both"
	logger.Initialize()
	assert.NotNil(t, logger.Log, "Logger should be initialized")
}

// TestInfo tests the Info function by checking the access log file
func TestInfo(t *testing.T) {
	tempDir := setupFileLogger(t)
	defer os.RemoveAll(tempDir)

	logger.Info("Test info message")

	content := readLogFile(t, tempDir, "access.log")
	assert.Contains(t, content, "Test info message", "Access log should contain the info message")
}

// TestInfof tests the Infof function
func TestInfof(t *testing.T) {
	tempDir := setupFileLogger(t)
	defer os.RemoveAll(tempDir)

	logger.Infof("Test info message with %s", "formatting")

	content := readLogFile(t, tempDir, "access.log")
	assert.Contains(t, content, "Test info message with formatting", "Access log should contain the formatted info message")
}

// TestError tests the Error function by checking the error log file
func TestError(t *testing.T) {
	tempDir := setupFileLogger(t)
	defer os.RemoveAll(tempDir)

	logger.Error("Test error message")

	content := readLogFile(t, tempDir, "error.log")
	assert.Contains(t, content, "Test error message", "Error log should contain the error message")
}

// TestErrorf tests the Errorf function
func TestErrorf(t *testing.T) {
	tempDir := setupFileLogger(t)
	defer os.RemoveAll(tempDir)

	logger.Errorf("Test error message with %s", "formatting")

	content := readLogFile(t, tempDir, "error.log")
	assert.Contains(t, content, "Test error message with formatting", "Error log should contain the formatted error message")
}

// TestWarning tests the Warning function by checking the error log file
func TestWarning(t *testing.T) {
	tempDir := setupFileLogger(t)
	defer os.RemoveAll(tempDir)

	logger.Warning("Test warning message")

	content := readLogFile(t, tempDir, "error.log")
	assert.Contains(t, content, "Test warning message", "Error log should contain the warning message")
}

// TestWarningf tests the Warningf function
func TestWarningf(t *testing.T) {
	tempDir := setupFileLogger(t)
	defer os.RemoveAll(tempDir)

	logger.Warningf("Test warning message with %s", "formatting")

	content := readLogFile(t, tempDir, "error.log")
	assert.Contains(t, content, "Test warning message with formatting", "Error log should contain the formatted warning message")
}

// TestDebug tests that Debug does not panic. Debug logs go to "debug" log
// which is not configured by default, so messages are silently dropped.
func TestDebug(t *testing.T) {
	tempDir := setupFileLogger(t)
	defer os.RemoveAll(tempDir)

	// Set debug level to 1
	originalDebugLevel := conf.DEBUG_LEVEL
	defer func() { conf.DEBUG_LEVEL = originalDebugLevel }()
	conf.DEBUG_LEVEL = 1

	// Should not panic even though "debug" log is not configured
	assert.NotPanics(t, func() {
		logger.Debug(1, "Test debug message level 1")
	}, "Debug should not panic")

	// Debug with level above threshold should be silently dropped
	assert.NotPanics(t, func() {
		logger.Debug(2, "Test debug message level 2")
	}, "Debug with high level should not panic")
}

// TestDebugf tests that Debugf does not panic
func TestDebugf(t *testing.T) {
	tempDir := setupFileLogger(t)
	defer os.RemoveAll(tempDir)

	originalDebugLevel := conf.DEBUG_LEVEL
	defer func() { conf.DEBUG_LEVEL = originalDebugLevel }()
	conf.DEBUG_LEVEL = 1

	assert.NotPanics(t, func() {
		logger.Debugf(1, "Test debug message level 1 with %s", "formatting")
	}, "Debugf should not panic")

	assert.NotPanics(t, func() {
		logger.Debugf(2, "Test debug message level 2 with %s", "formatting")
	}, "Debugf with high level should not panic")
}

// TestTrace tests the Trace function
func TestTrace(t *testing.T) {
	tempDir := setupFileLogger(t)
	defer os.RemoveAll(tempDir)

	originalLogTrace := conf.LOG_TRACE
	defer func() { conf.LOG_TRACE = originalLogTrace }()

	// With LOG_TRACE = false, Trace should not log
	conf.LOG_TRACE = false
	logger.Trace("Test trace message with LOG_TRACE = false")

	content := readLogFile(t, tempDir, "access.log")
	assert.NotContains(t, content, "Test trace message with LOG_TRACE = false",
		"Trace should not log when LOG_TRACE is false")

	// With LOG_TRACE = true, Trace should log to access
	conf.LOG_TRACE = true
	logger.Trace("Test trace message with LOG_TRACE = true")

	content = readLogFile(t, tempDir, "access.log")
	assert.Contains(t, content, "Test trace message with LOG_TRACE = true",
		"Trace should log when LOG_TRACE is true")
}

// TestTracef tests the Tracef function
func TestTracef(t *testing.T) {
	tempDir := setupFileLogger(t)
	defer os.RemoveAll(tempDir)

	originalLogTrace := conf.LOG_TRACE
	defer func() { conf.LOG_TRACE = originalLogTrace }()

	// With LOG_TRACE = false, Tracef should not log
	conf.LOG_TRACE = false
	logger.Tracef("Test trace message with LOG_TRACE = false and %s", "formatting")

	content := readLogFile(t, tempDir, "access.log")
	assert.NotContains(t, content, "Test trace message with LOG_TRACE = false and formatting",
		"Tracef should not log when LOG_TRACE is false")

	// With LOG_TRACE = true, Tracef should log to access
	conf.LOG_TRACE = true
	logger.Tracef("Test trace message with LOG_TRACE = true and %s", "formatting")

	content = readLogFile(t, tempDir, "access.log")
	assert.Contains(t, content, "Test trace message with LOG_TRACE = true and formatting",
		"Tracef should log when LOG_TRACE is true")
}

// TestPerf tests the Perf function
func TestPerf(t *testing.T) {
	tempDir := setupFileLogger(t)
	defer os.RemoveAll(tempDir)

	logger.Perf("Test perf message")

	content := readLogFile(t, tempDir, "perf.log")
	assert.Contains(t, content, "Test perf message", "Perf log should contain the perf message")
}
