package main

import (
	"errors"
	"math"

	"github.com/rs/zerolog/log"
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

	// ProbabilityTable is the table of pattern probabilities given the previous week probability
	ProbabilityTable = [][]float64{
		// {Random, BigSpike, Falling, SmallSpike}
		{20, 30, 15, 35}, // Random
		{50, 5, 20, 25},  // BigSpike
		{25, 45, 5, 25},  // Falling
		{45, 25, 15, 15}, // SmallSpike
	}
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

	Patterns      []Pattern
	MaxMin        [12]DayPrice
	Probabilities map[PatternType]float64
}

// NewForecast returns a Forecasts
func NewForecast(sellPrice uint32, buyPrices [12]uint32, previousWeek *Forecast) (*Forecast, error) {
	if sellPrice < 90 || sellPrice > 110 {
		return nil, ErrSellPrice
	}

	for _, price := range buyPrices {
		if price > 660 {
			return nil, ErrBuyPrice
		}
	}

	f := Forecast{
		sellPrice: sellPrice,
		buyPrices: buyPrices,
	}

	f.runForecast(previousWeek)

	return &f, nil
}

// Common operations

// minRate returns the minimum rate vs the sell price for a half day
func (f *Forecast) minRate(i uint8) float64 {
	return (float64(f.buyPrices[i]) - 0.5) / float64(f.sellPrice)
}

// maxRate returns the maximum rate vs the sell price for a half day
func (f *Forecast) maxRate(i uint8) float64 {
	return (float64(f.buyPrices[i]) + 0.5) / float64(f.sellPrice)
}

// minRatePrice applies a rate to the sell price and returns the result rounded to the lower value
func (f *Forecast) minRatePrice(r float64) uint32 {
	return uint32(math.Floor(r * float64(f.sellPrice)))
}

// maxRatePrice applies a rate to the sell price and returns the result rounded to the upper value
func (f *Forecast) maxRatePrice(r float64) uint32 {
	return uint32(math.Ceil(r * float64(f.sellPrice)))
}

// Random pattern functions

// genRandomPattern returns a random pattern following the given halfs for each phase
func (f *Forecast) genRandomPattern(inc1Halfs, dec1Halfs, inc2Halfs, dec2Halfs, inc3Halfs uint8) (*Pattern, error) {
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

// genRandomPatterns generates all possible random patterns with the current data
func (f *Forecast) genRandomPatterns() {
	for dec1 := uint8(2); dec1 < 4; dec1++ {
		for inc1 := uint8(0); inc1 < 7; inc1++ {
			for inc3 := uint8(0); inc3 < (7 - inc1); inc3++ {
				pattern, err := f.genRandomPattern(inc1, dec1, 7-inc1-inc3, 5-dec1, inc3)

				if err == nil {
					f.Patterns = append(f.Patterns, *pattern)
				} else if err != ErrNoMatch {
					log.Error().Str("module", "forecast").Err(err).Msg("error generating random pattern")
				}
			}
		}
	}
}

// genBigSpikePattern returns a big peak pattern given a spike start
func (f *Forecast) genBigSpikePattern(spikeStart uint8) (*Pattern, error) {
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

// genBigSpikePatterns generates all possible big spike patterns with the current data
func (f *Forecast) genBigSpikePatterns() {
	for spikeStart := uint8(1); spikeStart < 8; spikeStart++ {
		pattern, err := f.genBigSpikePattern(spikeStart)

		if err == nil {
			f.Patterns = append(f.Patterns, *pattern)
		} else if err != ErrNoMatch {
			log.Error().Str("module", "forecast").Err(err).Msg("error generating big spike pattern")
		}
	}
}

// genFallingPattern returns falling pattern
func (f *Forecast) genFallingPattern() (*Pattern, error) {
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

// genFallingPatterns generates all possible big spike patterns with the current data
func (f *Forecast) genFallingPatterns() {
	pattern, err := f.genFallingPattern()

	if err == nil {
		f.Patterns = append(f.Patterns, *pattern)
	} else if err != ErrNoMatch {
		log.Error().Str("module", "forecast").Err(err).Msg("error generating falling pattern")
	}
}

// genSmallSpikePattern returns a small peak pattern given a spike start
func (f *Forecast) genSmallSpikePattern(spikeStart uint8) (*Pattern, error) {
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

// genSmallSpikePatterns generates all possible small spike patterns with the current data
func (f *Forecast) genSmallSpikePatterns() {
	for spikeStart := uint8(0); spikeStart < 8; spikeStart++ {
		pattern, err := f.genSmallSpikePattern(spikeStart)

		if err == nil {
			f.Patterns = append(f.Patterns, *pattern)
		} else if err != ErrNoMatch {
			log.Error().Str("module", "forecast").Err(err).Msg("error generating small spike pattern")
		}
	}
}

// runForecast generates all patterns that could match and gets the max and min possible price per day
func (f *Forecast) runForecast(previousWeek *Forecast) {
	// Make sure everything is empty
	f.Patterns = []Pattern{}
	f.MaxMin = [12]DayPrice{}
	f.Probabilities = map[PatternType]float64{}

	// Generate all patterns
	f.genRandomPatterns()
	f.genBigSpikePatterns()
	f.genFallingPatterns()
	f.genSmallSpikePatterns()

	// Calc max and min values for each day and keep track possible pattern types
	patterns := map[PatternType]bool{}

	for i := range f.MaxMin {
		f.MaxMin[i].Min = math.MaxUint32
	}

	for i := range f.Patterns {
		patterns[f.Patterns[i].Type] = true

		for j := range f.Patterns[i].Prices {
			f.MaxMin[j].Max = maxUint32(f.MaxMin[j].Max, f.Patterns[i].Prices[j].Max)
			f.MaxMin[j].Min = minUint32(f.MaxMin[j].Min, f.Patterns[i].Prices[j].Min)
		}
	}

	// Calc probabilities
	if previousWeek != nil {
		for pattern := range patterns {
			for prevPattern, prevProb := range previousWeek.Probabilities {
				f.Probabilities[pattern] += (ProbabilityTable[prevPattern][pattern] * prevProb)
			}
		}
	} else {
		for pattern := range patterns {
			for _, patProb := range ProbabilityTable {
				f.Probabilities[pattern] += patProb[pattern]
			}
		}
	}

	// Normalize probabilities
	var total float64 = 0

	for _, p := range f.Probabilities {
		total += p
	}

	for p := range f.Probabilities {
		f.Probabilities[p] /= total
	}
}
