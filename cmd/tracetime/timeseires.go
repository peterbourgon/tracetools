package main

import (
	"fmt"
	"github.com/wcharczuk/go-chart"
	"os"
)

type plotSeries struct {
	stackID          string
	xvalues, yvalues []float64
}

func newTimeseries(seriesByFunc []plotSeries, funcName string) error {
	var allSeries []chart.Series

	for _, s := range seriesByFunc {
		allSeries = append(allSeries, chart.ContinuousSeries{
			Name: s.stackID,
			Style: chart.Style{
				Show:        true,
				StrokeColor: chart.ColorBlue,
				FillColor:   chart.ColorBlue.WithAlpha(100),
			},
			XValues: s.xvalues,
			YValues: s.yvalues,
		})
	}

	graph := chart.Chart{
		// Y axis
		YAxis: chart.YAxis{
			Name:      "Latency (ms)",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
			TickStyle: chart.Style{
				TextRotationDegrees: 45.0,
			},
			ValueFormatter: func(v interface{}) string {
				return fmt.Sprintf("%d ms", int(v.(float64)))
			},
		},
		// X axis
		XAxis: chart.XAxis{
			Style: chart.StyleShow(),
			TickStyle: chart.Style{
				TextRotationDegrees: 45.0,
			},
			ValueFormatter: func(v interface{}) string {
				return fmt.Sprintf("%d ms", int(v.(float64)))
			},
			GridMajorStyle: chart.Style{
				Show:        true,
				StrokeColor: chart.ColorAlternateGray,
				StrokeWidth: 1.0,
			},
		},
		Series: allSeries,
	}
	graph.Elements = []chart.Renderable{chart.LegendThin(&graph)}

	f, err := os.Create(funcName + ".png")
	if err != nil {
		return err
	}
	defer f.Close()
	graph.Render(chart.PNG, f)
	return nil
}
