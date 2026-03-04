package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/trace"
	"sort"
	"strings"
	"time"

	"github.com/MG-RAST/Shock/shock-server/conf"
)

var traceFile *os.File
var traceOn bool

func traceFileName() string {
	return fmt.Sprintf("trace.%d.log", time.Now().Unix())
}

// latestTraceFile returns the path to the most recent completed trace file.
func latestTraceFile() (string, error) {
	entries, err := os.ReadDir(conf.PATH_LOGS)
	if err != nil {
		return "", fmt.Errorf("cannot read logs directory: %w", err)
	}
	var traces []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "trace.") && strings.HasSuffix(e.Name(), ".log") {
			// Skip the currently-active trace file
			if traceOn && traceFile != nil && e.Name() == filepath.Base(traceFile.Name()) {
				continue
			}
			traces = append(traces, e.Name())
		}
	}
	if len(traces) == 0 {
		return "", fmt.Errorf("no trace files found")
	}
	sort.Strings(traces)
	return filepath.Join(conf.PATH_LOGS, traces[len(traces)-1]), nil
}

// runGoToolTrace runs "go tool trace -d=<mode> <file>" and returns the output.
func runGoToolTrace(traceFilePath string, mode string) ([]byte, error) {
	goPath, err := exec.LookPath("go")
	if err != nil {
		return nil, fmt.Errorf("go toolchain not found on PATH: %w", err)
	}
	cmd := exec.Command(goPath, "tool", "trace", "-d="+mode, traceFilePath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("go tool trace failed: %w\n%s", err, string(out))
	}
	return out, nil
}

func hourlyTrace() {
	wait := 60 * time.Minute
	for {
		durationTrace(wait)
	}
}

func minuteTrace() {
	wait := 1 * time.Minute
	for {
		durationTrace(wait)
	}
}

func durationTrace(wait time.Duration) {
	name := traceFileName()
	startTrace(name)
	defer stopTrace()
	time.Sleep(wait)
}

func startTrace(name string) (err error) {
	if traceOn && (traceFile != nil) {
		err = fmt.Errorf("tracing is already enabled with file %s", traceFile.Name())
		return
	}
	traceFile, err = os.Create(fmt.Sprintf("%s/%s", conf.PATH_LOGS, name))
	if err != nil {
		return
	}
	err = trace.Start(traceFile)
	if err != nil {
		traceFile.Close()
		return
	}
	traceOn = true
	return
}

func stopTrace() (err error) {
	if traceOn {
		trace.Stop()
		traceOn = false
		err = traceFile.Close()
	} else {
		err = fmt.Errorf("tracing is not enabled")
	}
	return
}
