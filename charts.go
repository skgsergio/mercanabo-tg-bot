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
	"time"

	"github.com/wcharczuk/go-chart"

	"github.com/rs/zerolog/log"
)

// PricesChart returns a chart given a slice of prices
func PricesChart(title string, xValues *[12]time.Time, yValues *[12]float64, lineValue float64, prediction *[12]DayPrice, location *time.Location, addRangeTitle bool) (*bytes.Buffer, error) {
	if addRangeTitle {
		title += fmt.Sprintf(" | %s - %s", xValues[0].Format("2006-01-02"), xValues[len(xValues)-1].Format("2006-01-02"))
	}

	xxValues := []time.Time{}
	yyValues := []float64{}

	for i := range yValues {
		if yValues[i] != 0 {
			xxValues = append(xxValues, xValues[i])
			yyValues = append(yyValues, yValues[i])
		}
	}

	// Graph series slice
	graphSeries := []chart.Series{}

	// Create price series
	priceSeries := chart.TimeSeries{
		Style: chart.Style{
			StrokeColor: chart.ColorBlue,
			FillColor:   chart.ColorTransparent,
		},
		XValues: xxValues,
		YValues: yyValues,
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

	// Create ticks for x axis
	ticks := make([]chart.Tick, len(xValues))

	// Fill annotations and ticks
	for i := 0; i < len(xValues); i++ {
		ticks[i].Value = chart.TimeToFloat64(xValues[i])
		ticks[i].Label = TimeToShortDayAMPM(xValues[i])
	}

	graphSeries = append(graphSeries, priceAnnotations)

	// Create prediction series if any
	if prediction != nil {
		predMinSeries := chart.TimeSeries{
			Style: chart.Style{
				StrokeColor:     chart.ColorOrange,
				StrokeDashArray: []float64{5.0, 5.0},
				FillColor:       chart.ColorTransparent,
			},
			XValues: xValues[:],
			YValues: []float64{},
		}

		predMaxSeries := chart.TimeSeries{
			Style: chart.Style{
				StrokeColor:     chart.ColorOrange,
				StrokeDashArray: []float64{5.0, 5.0},
				FillColor:       chart.ColorTransparent,
			},
			XValues: xValues[:],
			YValues: []float64{},
		}

		for _, v := range prediction {
			predMinSeries.YValues = append(predMinSeries.YValues, float64(v.Min))
			predMaxSeries.YValues = append(predMaxSeries.YValues, float64(v.Max))
		}

		graphSeries = append(graphSeries, predMinSeries)
		graphSeries = append(graphSeries, predMaxSeries)
	}

	// Create owned series if the lineValue is not 0
	if lineValue != 0 {
		// Dashed line marking buy price
		ownedSeries := chart.TimeSeries{
			Style: chart.Style{
				StrokeColor:     chart.ColorRed,
				StrokeDashArray: []float64{5.0, 5.0},
			},
			XValues: xValues[:],
			YValues: make([]float64, len(xValues)),
		}

		for i := range ownedSeries.XValues {
			ownedSeries.YValues[i] = lineValue
		}

		graphSeries = append(graphSeries, ownedSeries)

		// Annotate buy price
		ownedAnnotation := chart.LastValueAnnotationSeries(ownedSeries)

		graphSeries = append(graphSeries, ownedAnnotation)
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
		XAxis: chart.XAxis{
			Ticks: ticks,
			GridMinorStyle: chart.Style{
				StrokeColor:     chart.ColorAlternateGray.WithAlpha(128),
				StrokeWidth:     1.0,
				StrokeDashArray: []float64{5.0, 5.0},
			},
		},
		YAxis: chart.YAxis{
			Name: texts.Bells,
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

// ChartValueToTime casts a chart value to a time.Time if possible
func ChartValueToTime(v interface{}, location *time.Location) time.Time {
	if t, ok := v.(time.Time); ok {
		return t.In(location)
	}

	if i, ok := v.(int64); ok {
		return time.Unix(0, i).In(location)
	}

	if f, ok := v.(float64); ok {
		return time.Unix(0, int64(f)).In(location)
	}

	return time.Time{}
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
