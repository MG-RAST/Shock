package main

import (
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"os"
	"runtime/trace"
	"time"
)

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
	epoc := time.Now().Unix()

	traceFile, _ := os.Create(fmt.Sprintf("%s/trace.%d.log", conf.PATH_LOGS, epoc))
	defer traceFile.Close()

	trace.Start(traceFile)
	defer trace.Stop()

	time.Sleep(wait)
}
