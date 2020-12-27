package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/peterbourgon/tracetools/pprof"
	"github.com/peterbourgon/tracetools/trace"
)

var usageMessage = `Usage: tracetime -filter=ServeHTTP [binary] trace.out

 -filter=REGEX   Function events matching this regex will be plotted.`

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

	//Extract the time series from the list events
	seriesMap := extractTimeSeries(events, re)

	//Group timeseries by function name
	seriesPerFunc := make(map[string][]timeSeries)
	for _, series := range seriesMap {
		s := seriesPerFunc[series.fnName]
		s = append(s, series)
		seriesPerFunc[series.fnName] = s
	}

	//Plot overlapping graph for each function
	for fnName, setOfSeries := range seriesPerFunc {
		var allSeries []plotSeries
		for _, s := range setOfSeries {
			var p plotSeries
			for _, lat := range s.durations {
				p.yvalues = append(p.yvalues, float64(lat.duration))
				p.xvalues = append(p.xvalues, float64(lat.ts))
			}
			p.stackID = fmt.Sprintf("Stack %v", s.stkID)
			allSeries = append(allSeries, p)
		}
		newTimeseries(allSeries, fnName)
	}
}

type latency struct {
	ts       int64 // First timestamp when the function was called
	duration int64 // How long the function took
}

type timeSeries struct {
	fnName    string
	stkID     uint64
	durations []latency
}

func extractTimeSeries(events []*trace.Event, regex *regexp.Regexp) map[uint64]timeSeries {
	series := make(map[uint64]timeSeries)
	for _, ev := range events {
		if ev.Link == nil || ev.StkID == 0 || len(ev.Stk) == 0 {
			continue
		}
		if regex.FindStringIndex(ev.Stk[0].Fn) == nil {
			continue
		}
		d := series[ev.StkID]
		d.stkID = ev.StkID
		d.durations = append(d.durations, latency{
			ts:       ev.Ts,
			duration: ev.Link.Ts - ev.Ts,
		})
		d.fnName = ev.Stk[0].Fn
		series[ev.StkID] = d
	}
	return series
}
