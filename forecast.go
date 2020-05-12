package main

import (
	"errors"
	"fmt"
	"math"
)

var (
	// ErrNoMatch is returned when the generated pattern doesn't match current prices
	ErrNoMatch = errors.New("pattern doesn't match")
)

// PatternType is the type of price patterns
type PatternType int

const (
	// Random means that prices will be random. Max price 1.1x~1.45x
	Random PatternType = iota
	// BigSpike means that a big spike will occur. 2nd interval > 1.4x, 3rd interval 2x~6x
	BigSpike
	// Falling means that prices will continue falling. 100% lose out
	Falling
	// SmallSpike means that a small spike will occur. 2nd interval < 1.4x, 4th interval 1.4x~2x
	SmallSpike
)

var (
	// ErrSellPrice is returned when a new forecast is initializated and the sell price is less than 90 or greater than 110
	ErrSellPrice = errors.New("sell price can't be lower than 90 or greater than 110")
	// ErrBuyPrice is returned when a new forecast is initializated and the buy price greater than 660
	ErrBuyPrice = errors.New("buy price can't be greater than 660")
)

// DayPrice represents the max and min prices in a day
type DayPrice struct {
	Min uint32
	Max uint32
}

// Pattern represents a pattern following the ACNH algorithm
type Pattern struct {
	Type   PatternType
	Prices [12]DayPrice
}

// Forecast represents a predicion of the Stalk Market
type Forecast struct {
	sellPrice uint32
	buyPrices [12]uint32
}

// NewForecast returns a Forecasts
func NewForecast(sellPrice uint32, buyPrices [12]uint32) (*Forecast, error) {
	if sellPrice < 90 || sellPrice > 110 {
		return nil, ErrSellPrice
	}

	for _, price := range buyPrices {
		if price > 660 {
			return nil, ErrBuyPrice
		}
	}

	return &Forecast{
		sellPrice: sellPrice,
		buyPrices: buyPrices,
	}, nil
}

// Common operations

func (f *Forecast) minRate(i uint8) float64 {
	return (float64(f.buyPrices[i]) - 0.5) / float64(f.sellPrice)
}

func (f *Forecast) maxRate(i uint8) float64 {
	return (float64(f.buyPrices[i]) + 0.5) / float64(f.sellPrice)
}

func (f *Forecast) minRatePrice(r float64) uint32 {
	return uint32(math.Floor(r * float64(f.sellPrice)))
}

func (f *Forecast) maxRatePrice(r float64) uint32 {
	return uint32(math.Ceil(r * float64(f.sellPrice)))
}

// Random pattern functions

// GenRandomPattern returns a random pattern following the given halfs for each phase
func (f *Forecast) GenRandomPattern(inc1Halfs, dec1Halfs, inc2Halfs, dec2Halfs, inc3Halfs uint8) (*Pattern, error) {
	pattern := Pattern{
		Type: Random,
	}

	// Check constraints
	if inc1Halfs > 6 {
		return nil, errors.New("increase phase 1 must be between 0 and 6 half days")
	}

	if dec1Halfs != 2 && dec1Halfs != 3 {
		return nil, errors.New("decrease phase 1 must be 2 or 3 half days")
	}

	if 7-inc1Halfs-inc3Halfs != inc2Halfs {
		return nil, errors.New("increase phase 2 must be 7 half says - increase phase 1 - increase phase 3")
	}

	if 5-dec1Halfs != dec2Halfs {
		return nil, errors.New("decrease phase 2 must be 5 half days - decrease phase 1")
	}

	if inc3Halfs >= 7-inc1Halfs {
		return nil, errors.New("increase phase 3 must be between 0 and 7 half says - increase phase 1 - 1")
	}

	if inc1Halfs+dec1Halfs+inc2Halfs+dec2Halfs+inc3Halfs != 12 {
		return nil, errors.New("phase halfs must sum 12")
	}

	// Increase Phase 1
	halfs := inc1Halfs
	for i := halfs - inc1Halfs; i < halfs; i++ {
		minPred := f.minRatePrice(0.9)
		maxPred := f.maxRatePrice(1.4)

		if f.buyPrices[i] != 0 {
			if f.buyPrices[i] < minPred || f.buyPrices[i] > maxPred {
				return nil, ErrNoMatch
			}

			minPred = f.buyPrices[i]
			maxPred = f.buyPrices[i]
		}

		pattern.Prices[i] = DayPrice{
			Min: minPred,
			Max: maxPred,
		}
	}

	// Decrease Phase 1
	minRate := 0.6
	maxRate := 0.8

	halfs += dec1Halfs
	for i := halfs - dec1Halfs; i < halfs; i++ {
		minPred := f.minRatePrice(minRate)
		maxPred := f.maxRatePrice(maxRate)

		if f.buyPrices[i] != 0 {
			if f.buyPrices[i] < minPred || f.buyPrices[i] > maxPred {
				return nil, ErrNoMatch
			}

			minPred = f.buyPrices[i]
			maxPred = f.buyPrices[i]

			minRate = f.minRate(i)
			maxRate = f.maxRate(i)
		}

		pattern.Prices[i] = DayPrice{
			Min: minPred,
			Max: maxPred,
		}

		minRate -= 0.1
		maxRate -= 0.04
	}

	// Increase Phase 2
	halfs += inc2Halfs
	for i := halfs - inc2Halfs; i < halfs; i++ {
		minPred := f.minRatePrice(0.9)
		maxPred := f.maxRatePrice(1.4)

		if f.buyPrices[i] != 0 {
			if f.buyPrices[i] < minPred || f.buyPrices[i] > maxPred {
				return nil, ErrNoMatch
			}

			minPred = f.buyPrices[i]
			maxPred = f.buyPrices[i]
		}

		pattern.Prices[i] = DayPrice{
			Min: minPred,
			Max: maxPred,
		}
	}

	// Decrease Phase 2
	minRate = 0.6
	maxRate = 0.8

	halfs += dec2Halfs
	for i := halfs - dec2Halfs; i < halfs; i++ {
		minPred := f.minRatePrice(minRate)
		maxPred := f.maxRatePrice(maxRate)

		if f.buyPrices[i] != 0 {
			if f.buyPrices[i] < minPred || f.buyPrices[i] > maxPred {
				return nil, ErrNoMatch
			}

			minPred = f.buyPrices[i]
			maxPred = f.buyPrices[i]

			minRate = f.minRate(i)
			maxRate = f.maxRate(i)
		}

		pattern.Prices[i] = DayPrice{
			Min: minPred,
			Max: maxPred,
		}

		minRate -= 0.1
		maxRate -= 0.04
	}

	// Increase Phase 3
	halfs += inc3Halfs
	for i := halfs - inc3Halfs; i < halfs; i++ {
		minPred := f.minRatePrice(0.9)
		maxPred := f.maxRatePrice(1.4)

		if f.buyPrices[i] != 0 {
			if f.buyPrices[i] < minPred || f.buyPrices[i] > maxPred {
				return nil, ErrNoMatch
			}

			minPred = f.buyPrices[i]
			maxPred = f.buyPrices[i]
		}

		pattern.Prices[i] = DayPrice{
			Min: minPred,
			Max: maxPred,
		}
	}

	return &pattern, nil
}

// GenRandomPatterns generates all possible random patterns with the current data
func (f *Forecast) GenRandomPatterns() []*Pattern {
	patterns := []*Pattern{}

	for dec1 := uint8(2); dec1 < 4; dec1++ {
		for inc1 := uint8(0); inc1 < 7; inc1++ {
			for inc3 := uint8(0); inc3 < (7 - inc1); inc3++ {
				pattern, err := f.GenRandomPattern(inc1, dec1, 7-inc1-inc3, 5-dec1, inc3)

				if err == nil {
					patterns = append(patterns, pattern)
				} else if err != ErrNoMatch {
					fmt.Printf("Error: %v\n", err)
				}
			}
		}
	}

	return patterns
}

// GenBigSpikePattern returns a big peak pattern given a spike start
func (f *Forecast) GenBigSpikePattern(spikeStart uint8) (*Pattern, error) {
	pattern := Pattern{
		Type: BigSpike,
	}

	// Check constraints
	if spikeStart < 1 || spikeStart > 7 {
		return nil, errors.New("spike start must be between 1 and 7")
	}

	// Decrease Phase
	minRate := 0.85
	maxRate := 0.9

	for i := uint8(0); i < spikeStart; i++ {
		minPred := f.minRatePrice(minRate)
		maxPred := f.maxRatePrice(maxRate)

		if f.buyPrices[i] != 0 {
			if f.buyPrices[i] < minPred || f.buyPrices[i] > maxPred {
				return nil, ErrNoMatch
			}

			minPred = f.buyPrices[i]
			maxPred = f.buyPrices[i]

			minRate = f.minRate(i)
			maxRate = f.maxRate(i)
		}

		pattern.Prices[i] = DayPrice{
			Min: minPred,
			Max: maxPred,
		}

		minRate -= 0.05
		maxRate -= 0.03
	}

	// Sharp Increase Phase + Sharp Decrease Phase
	minRates := []float64{0.9, 1.4, 2.0, 1.4, 0.9}
	maxRates := []float64{1.4, 2.0, 6.0, 2.0, 1.4}

	for i := spikeStart; i < spikeStart+5; i++ {
		minPred := f.minRatePrice(minRates[i-spikeStart])
		maxPred := f.maxRatePrice(maxRates[i-spikeStart])

		if f.buyPrices[i] != 0 {
			if f.buyPrices[i] < minPred || f.buyPrices[i] > maxPred {
				return nil, ErrNoMatch
			}

			minPred = f.buyPrices[i]
			maxPred = f.buyPrices[i]

		}

		pattern.Prices[i] = DayPrice{
			Min: minPred,
			Max: maxPred,
		}
	}

	// Random decrease
	for i := spikeStart + 5; i < 12; i++ {
		minPred := f.minRatePrice(0.4)
		maxPred := f.maxRatePrice(0.9)

		if f.buyPrices[i] != 0 {
			if f.buyPrices[i] < minPred || f.buyPrices[i] > maxPred {
				return nil, ErrNoMatch
			}

			minPred = f.buyPrices[i]
			maxPred = f.buyPrices[i]

		}

		pattern.Prices[i] = DayPrice{
			Min: minPred,
			Max: maxPred,
		}
	}

	return &pattern, nil
}

// GenBigSpikePatterns generates all possible big spike patterns with the current data
func (f *Forecast) GenBigSpikePatterns() []*Pattern {
	patterns := []*Pattern{}

	for spikeStart := uint8(1); spikeStart < 8; spikeStart++ {
		pattern, err := f.GenBigSpikePattern(spikeStart)

		if err == nil {
			patterns = append(patterns, pattern)
		} else if err != ErrNoMatch {
			fmt.Printf("Error: %v\n", err)
		}
	}

	return patterns
}

// GenFallingPattern returns falling pattern
func (f *Forecast) GenFallingPattern() (*Pattern, error) {
	pattern := Pattern{
		Type: Falling,
	}

	minRate := 0.85
	maxRate := 0.9

	for i := uint8(0); i < 12; i++ {
		minPred := f.minRatePrice(minRate)
		maxPred := f.maxRatePrice(maxRate)

		if f.buyPrices[i] != 0 {
			if f.buyPrices[i] < minPred || f.buyPrices[i] > maxPred {
				return nil, ErrNoMatch
			}

			minPred = f.buyPrices[i]
			maxPred = f.buyPrices[i]

			minRate = f.minRate(i)
			maxRate = f.maxRate(i)
		}

		pattern.Prices[i] = DayPrice{
			Min: minPred,
			Max: maxPred,
		}

		minRate -= 0.05
		maxRate -= 0.03
	}

	return &pattern, nil
}

// GenFallingPatterns generates all possible big spike patterns with the current data
func (f *Forecast) GenFallingPatterns() []*Pattern {
	patterns := []*Pattern{}

	pattern, err := f.GenFallingPattern()

	if err == nil {
		patterns = append(patterns, pattern)
	} else if err != ErrNoMatch {
		fmt.Printf("Error: %v\n", err)
	}

	return patterns
}

// GenSmallSpikePattern returns a small peak pattern given a spike start
func (f *Forecast) GenSmallSpikePattern(spikeStart uint8) (*Pattern, error) {
	pattern := Pattern{
		Type: SmallSpike,
	}

	// Check constraints
	if spikeStart > 7 {
		return nil, errors.New("spike start must be between 1 and 7")
	}

	// Decrease Phase
	minRate := 0.4
	maxRate := 0.9

	for i := uint8(0); i < spikeStart; i++ {
		minPred := f.minRatePrice(minRate)
		maxPred := f.maxRatePrice(maxRate)

		if f.buyPrices[i] != 0 {
			if f.buyPrices[i] < minPred || f.buyPrices[i] > maxPred {
				return nil, ErrNoMatch
			}

			minPred = f.buyPrices[i]
			maxPred = f.buyPrices[i]

			minRate = f.minRate(i)
			maxRate = f.maxRate(i)
		}

		pattern.Prices[i] = DayPrice{
			Min: minPred,
			Max: maxPred,
		}

		minRate -= 0.05
		maxRate -= 0.03
	}

	// Increase Phase
	minRates := []float64{0.9, 0.9, 1.4, 1.4, 1.4}
	maxRates := []float64{1.4, 1.4, 2.0, 2.0, 2.0}

	for i := spikeStart; i < spikeStart+5; i++ {
		minPred := f.minRatePrice(minRates[i-spikeStart])
		maxPred := f.maxRatePrice(maxRates[i-spikeStart])

		// 3rd and 5th days are locked to maxRate - 1
		if i-spikeStart == 2 || i-spikeStart == 4 {
			maxPred--
		}

		// 4th day is the peak where the minimum is the previous day minimum
		if i-spikeStart == 3 {
			minPred = pattern.Prices[i-1].Min
		}

		if f.buyPrices[i] != 0 {
			if f.buyPrices[i] < minPred || f.buyPrices[i] > maxPred {
				return nil, ErrNoMatch
			}

			minPred = f.buyPrices[i]
			maxPred = f.buyPrices[i]
		}

		pattern.Prices[i] = DayPrice{
			Min: minPred,
			Max: maxPred,
		}
	}

	// Decrease Phase 2
	minRate = 0.4
	maxRate = 0.9

	for i := spikeStart + 5; i < 12; i++ {
		minPred := f.minRatePrice(minRate)
		maxPred := f.maxRatePrice(maxRate)

		if f.buyPrices[i] != 0 {
			if f.buyPrices[i] < minPred || f.buyPrices[i] > maxPred {
				return nil, ErrNoMatch
			}

			minPred = f.buyPrices[i]
			maxPred = f.buyPrices[i]

			minRate = f.minRate(i)
			maxRate = f.maxRate(i)
		}

		pattern.Prices[i] = DayPrice{
			Min: minPred,
			Max: maxPred,
		}

		minRate -= 0.05
		maxRate -= 0.03
	}

	return &pattern, nil
}

// GenSmallSpikePatterns generates all possible small spike patterns with the current data
func (f *Forecast) GenSmallSpikePatterns() []*Pattern {
	patterns := []*Pattern{}

	for spikeStart := uint8(0); spikeStart < 8; spikeStart++ {
		pattern, err := f.GenSmallSpikePattern(spikeStart)

		if err == nil {
			patterns = append(patterns, pattern)
		} else if err != ErrNoMatch {
			fmt.Printf("Error: %v\n", err)
		}
	}

	return patterns
}

// GenAllPatterns generates all pattterns
func (f *Forecast) GenAllPatterns() (*[]*Pattern, *[12]DayPrice) {
	patterns := []*Pattern{}
	maxMin := [12]DayPrice{}

	patterns = append(patterns, f.GenRandomPatterns()...)
	patterns = append(patterns, f.GenBigSpikePatterns()...)
	patterns = append(patterns, f.GenFallingPatterns()...)
	patterns = append(patterns, f.GenSmallSpikePatterns()...)

	if len(patterns) == 0 {
		return &patterns, nil
	}

	for i := range maxMin {
		maxMin[i].Min = math.MaxUint32
	}

	for _, pat := range patterns {
		for i, p := range pat.Prices {
			maxMin[i].Max = maxUint32(maxMin[i].Max, p.Max)
			maxMin[i].Min = minUint32(maxMin[i].Min, p.Min)
		}
	}

	return &patterns, &maxMin
}
