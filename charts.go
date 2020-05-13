// Copyright (c) 2020 Sergio Conde skgsergio@gmail.com
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU General Public License as published by the Free Software
// Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
// PARTICULAR PURPOSE. See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along with
// this program. If not, see <https://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"bytes"
	"fmt"
	"math"
	"time"

	"github.com/wcharczuk/go-chart"

	"github.com/rs/zerolog/log"
)

// PricesChart returns a chart given a slice of prices
func PricesChart(title string, times *[12]time.Time, prices *[12]uint32, ownedBells uint32, forecast *Forecast, location *time.Location, addRangeTitle bool) (*bytes.Buffer, error) {
	if addRangeTitle {
		title += fmt.Sprintf(" | %s - %s", times[0].Format("2006-01-02"), times[len(times)-1].Format("2006-01-02"))
	}

	// Graph series slice
	graphSeries := []chart.Series{}

	// Create owned series if the ownedValue is not 0
	if ownedBells != 0 {
		// Dashed line marking buy price
		ownedSeries := chart.TimeSeries{
			Style: chart.Style{
				StrokeColor:     chart.ColorRed,
				StrokeDashArray: []float64{5.0, 5.0},
			},
			XValues: []time.Time{times[0], times[11]},
			YValues: []float64{float64(ownedBells), float64(ownedBells)},
		}

		graphSeries = append(graphSeries, ownedSeries)

		// Annotate buy price
		ownedAnnotation := chart.LastValueAnnotationSeries(ownedSeries)

		graphSeries = append(graphSeries, ownedAnnotation)
	}

	// Create prediction series if any
	if forecast != nil && len(forecast.Patterns) > 0 {
		predMinSeries := chart.TimeSeries{
			Style: chart.Style{
				StrokeColor:     chart.ColorOrange,
				StrokeDashArray: []float64{5.0, 5.0},
			},
			XValues: times[:],
			YValues: []float64{},
		}

		predMaxSeries := chart.TimeSeries{
			Style: chart.Style{
				StrokeColor:     chart.ColorOrange,
				StrokeDashArray: []float64{5.0, 5.0},
			},
			XValues: times[:],
			YValues: []float64{},
		}

		for _, v := range forecast.MaxMin {
			predMinSeries.YValues = append(predMinSeries.YValues, float64(v.Min))
			predMaxSeries.YValues = append(predMaxSeries.YValues, float64(v.Max))
		}

		graphSeries = append(graphSeries, predMinSeries)
		graphSeries = append(graphSeries, predMaxSeries)
	}

	// Create price series
	xValues := []time.Time{}
	yValues := []float64{}

	for i := range prices {
		if prices[i] != 0 {
			xValues = append(xValues, times[i])
			yValues = append(yValues, float64(prices[i]))
		}
	}

	priceSeries := chart.TimeSeries{
		Style: chart.Style{
			StrokeColor: chart.ColorBlue,
		},
		XValues: xValues,
		YValues: yValues,
	}

	graphSeries = append(graphSeries, priceSeries)

	// Create price series annotations
	priceAnnotations := chart.AnnotationSeries{
		Style: chart.Style{
			StrokeColor: priceSeries.Style.StrokeColor,
		},
		Annotations: []chart.Value2{},
	}

	for i := 0; i < priceSeries.Len(); i++ {
		x, y := priceSeries.GetValues(i)
		priceAnnotations.Annotations = append(
			priceAnnotations.Annotations,
			chart.Value2{
				XValue: x,
				YValue: y,
				Label:  fmt.Sprintf("%v", y),
			},
		)
	}

	graphSeries = append(graphSeries, priceAnnotations)

	// Ok, here is the deal: you walk away and act as if you didn't see this, and I explain to you this hack.
	//
	// When there is only one data point in the graph the library enters in an infinite loop state that is
	// related to the Y axis range generation. In order to avoid this we just create our own range for the
	// Y axis. Will report the bug.
	YRange := &chart.ContinuousRange{
		Max: -math.MaxFloat64,
		Min: math.MaxFloat64,
	}

	for _, price := range priceSeries.YValues {
		YRange.Max = math.Max(YRange.Max, price)
		YRange.Min = math.Min(YRange.Min, price)
	}

	if forecast != nil && len(forecast.Patterns) > 0 {
		for _, price := range forecast.MaxMin {
			YRange.Max = math.Max(YRange.Max, float64(price.Max))
			YRange.Min = math.Min(YRange.Min, float64(price.Min))
		}
	}

	YRange.Max += 5
	YRange.Min -= 5

	if YRange.Min < 0 {
		YRange.Min = 0
	}

	// Create ticks for x axis
	ticks := make([]chart.Tick, len(times))

	// Fill annotations and ticks
	for i := 0; i < len(times); i++ {
		ticks[i].Value = chart.TimeToFloat64(times[i])
		ticks[i].Label = TimeToShortDayAMPM(times[i])
	}

	// Create the graph
	graph := chart.Chart{
		Log:    &ZerologGoChart{},
		Width:  1280,
		Height: 720,
		DPI:    96,
		Title:  title,
		TitleStyle: chart.Style{
			FontSize: 12,
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top: 35,
			},
		},
		XAxis: chart.XAxis{
			Ticks: ticks,
			GridMinorStyle: chart.Style{
				StrokeColor:     chart.ColorAlternateGray.WithAlpha(128),
				StrokeWidth:     1.0,
				StrokeDashArray: []float64{5.0, 5.0},
			},
		},
		YAxis: chart.YAxis{
			Name:  texts.Bells,
			Range: YRange,
		},
		Series: graphSeries,
	}

	// Render the graph
	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		log.Error().Str("module", "chart").Err(err).Msg("failed rendering prices chart")
	}

	return buffer, err
}

// TimeToShortDayAMPM prints the name of the weekday plus AM or PM
func TimeToShortDayAMPM(t time.Time) string {
	return texts.DaysShort[t.Weekday()] + " " + t.Format("PM")
}

// ZerologGoChart is a simple custom logger using Zerolog for go-chart
type ZerologGoChart struct{}

// Info writes an info message.
func (l *ZerologGoChart) Info(arguments ...interface{}) {
	log.Info().Str("module", "chart").Interface("arguments", arguments).Msgf("go-chart info")
}

// Infof writes an info message.
func (l *ZerologGoChart) Infof(format string, arguments ...interface{}) {
	log.Info().Str("module", "chart").Msgf(format, arguments...)
}

// Debug writes an debug message.
func (l *ZerologGoChart) Debug(arguments ...interface{}) {
	log.Debug().Str("module", "chart").Interface("arguments", arguments).Msgf("go-chart debug")
}

// Debugf writes an debug message.
func (l *ZerologGoChart) Debugf(format string, arguments ...interface{}) {
	log.Debug().Str("module", "chart").Msgf(format, arguments...)
}

// Error writes an error message.
func (l *ZerologGoChart) Error(arguments ...interface{}) {
	log.Error().Str("module", "chart").Interface("arguments", arguments).Msgf("go-chart error")
}

// Errorf writes an error message.
func (l *ZerologGoChart) Errorf(format string, arguments ...interface{}) {
	log.Error().Str("module", "chart").Msgf(format, arguments...)
}

// Err writes an error message.
func (l *ZerologGoChart) Err(err error) {
	if err != nil {
		log.Error().Str("module", "chart").Err(err).Msg("go-chart err")
	}
}

// FatalErr writes an error message and exits.
func (l *ZerologGoChart) FatalErr(err error) {
	if err != nil {
		log.Fatal().Str("module", "chart").Err(err).Msg("go-chart fatal err")
	}
}
