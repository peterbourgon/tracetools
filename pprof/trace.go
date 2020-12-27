package pprof

import (
	"bufio"
	"fmt"
	"os"

	"github.com/peterbourgon/tracetools/trace"
)

func LoadTrace(traceFile, programBinary string) ([]*trace.Event, error) {
	tracef, err := os.Open(traceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open trace file: %v", err)
	}
	defer tracef.Close()

	// Parse and symbolize.
	parseRes, err := trace.Parse(bufio.NewReader(tracef), programBinary)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trace: %v", err)
	}

	return parseRes.Events, nil
}
