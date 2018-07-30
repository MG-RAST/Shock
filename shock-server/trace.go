package main

import (
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"os"
	"runtime/trace"
	"time"
)

var traceFile *os.File
var traceOn bool

func traceFileName() string {
	return fmt.Sprintf("trace.%d.log", time.Now().Unix())
}

func dailyTrace() {
	wait := 24 * time.Hour
	for {
		durationTrace(wait)
	}
}

func hourlyTrace() {
	wait := 60 * time.Minute
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

func stopTrace() {
	if traceOn {
		trace.Stop()
		traceOn = false
		traceFile.Close()
	}
}
