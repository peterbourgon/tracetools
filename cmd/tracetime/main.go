package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/FiloSottile/tracetools/pprof"
	"github.com/FiloSottile/tracetools/trace"
)

var usageMessage = `Usage: tracetime -filter=ServeHTTP [binary] trace.out

 -filter=REGEX   Syscall events matching this regex will be plotted.`

var (
	filter = flag.String("filter", "", "")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usageMessage)
		os.Exit(2)
	}
	flag.Parse()

	// Go 1.7 traces embed symbol info and does not require the binary.
	// But we optionally accept binary as first arg for Go 1.5 traces.
	var programBinary, traceFile string
	switch flag.NArg() {
	case 1:
		traceFile = flag.Arg(0)
	case 2:
		programBinary = flag.Arg(0)
		traceFile = flag.Arg(1)
	default:
		flag.Usage()
	}

	if *filter == "" {
		flag.Usage()
	}
	re, err := regexp.Compile(*filter)
	if err != nil {
		log.Fatalln("Failed to compile filter regex:", err)
	}

	events, err := pprof.LoadTrace(traceFile, programBinary)
	if err != nil {
		log.Fatal(err)
	}

	seriesMap := extractTimeSeries(events, re)
	for id, series := range seriesMap {
		log.Println("Series for stack ID =", id, "fn name =", series.fnName)
		var latencies, times []float64
		for _, lat := range series.durations {
			log.Printf("\tLatancy = %+v\n", lat)
			latencies = append(latencies, float64(lat.duration/1000000)) //From nanoseconds to milliseconds
			times = append(times, float64(lat.ts/1000000))
		}
		newTimeseries(times, latencies)
	}
}

type latency struct {
	ts       int64 // First timestamp when the function was called
	duration int64 // How long the function took
}

type timeSeries struct {
	fnName    string
	durations []latency
}

func extractTimeSeries(events []*trace.Event, regex *regexp.Regexp) map[uint64]timeSeries {
	series := make(map[uint64]timeSeries)
	for _, ev := range events {
		if ev.Type != trace.EvGoSysCall || ev.Link == nil || ev.StkID == 0 || len(ev.Stk) == 0 {
			continue
		}
		if regex.FindStringIndex(ev.Stk[0].Fn) == nil {
			continue
		}
		d := series[ev.StkID]
		d.durations = append(d.durations, latency{
			ts:       ev.Ts,
			duration: ev.Link.Ts - ev.Ts,
		})
		d.fnName = ev.Stk[0].Fn
		series[ev.StkID] = d
	}
	return series
}
