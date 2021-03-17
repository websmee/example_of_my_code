package app

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/advice"
	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
	"github.com/websmee/example_of_my_code/adviser/domain/params"
)

const optimizerThreads = 16

type ParamsOptimizerApp interface {
	OptimizeParams(ctx context.Context, name, quoteSymbol string, from, to time.Time, minFrequency float64) error
}

type optimizerApp struct {
	candlestickRepository candlestick.Repository
	paramsRepository      params.Repository
	adviser               advice.Adviser
	calc                  candlestick.Calculator
	startParams           []decimal.Decimal
	minParams             []decimal.Decimal
	maxParams             []decimal.Decimal
	modifyRate            float64
}

func NewCBSOptimizerApp(
	candlestickRepository candlestick.Repository,
	paramsRepository params.Repository,
	modifyRate float64,
) ParamsOptimizerApp {
	return newOptimizerApp(
		candlestickRepository,
		paramsRepository,
		advice.NewCBSAdviser(candlestickRepository, candlestick.DefaultCalculator()),
		minCBSParams().GetParams(),
		minCBSParams().GetParams(),
		maxCBSParams().GetParams(),
		modifyRate,
	)
}

func NewCBSROptimizerApp(
	candlestickRepository candlestick.Repository,
	paramsRepository params.Repository,
	modifyRate float64,
) ParamsOptimizerApp {
	return newOptimizerApp(
		candlestickRepository,
		paramsRepository,
		advice.NewCBSRAdviser(candlestickRepository, candlestick.DefaultCalculator()),
		minCBSRParams().GetParams(),
		minCBSRParams().GetParams(),
		maxCBSRParams().GetParams(),
		modifyRate,
	)
}

func NewFTOptimizerApp(
	candlestickRepository candlestick.Repository,
	paramsRepository params.Repository,
	modifyRate float64,
) ParamsOptimizerApp {
	return newOptimizerApp(
		candlestickRepository,
		paramsRepository,
		advice.NewFTAdviser(candlestickRepository, candlestick.DefaultCalculator()),
		minFTParams().GetParams(),
		minFTParams().GetParams(),
		maxFTParams().GetParams(),
		modifyRate,
	)
}

func newOptimizerApp(
	candlestickRepository candlestick.Repository,
	paramsRepository params.Repository,
	adviser advice.Adviser,
	startParams []decimal.Decimal,
	minParams []decimal.Decimal,
	maxParams []decimal.Decimal,
	modifyRate float64,
) ParamsOptimizerApp {
	candlestickRepository = candlestick.NewBasicFilter(candlestickRepository)
	return &optimizerApp{
		candlestickRepository: candlestickRepository,
		paramsRepository:      paramsRepository,
		adviser:               adviser,
		calc:                  candlestick.DefaultCalculator(),
		startParams:           startParams,
		minParams:             minParams,
		maxParams:             maxParams,
		modifyRate:            modifyRate,
	}
}

type paramsStats struct {
	params    []decimal.Decimal
	frequency float64
	accuracy  float64
}

func minCBSParams() *params.CBS {
	return &params.CBS{
		CalmDurationHours:   15,
		CalmMaxVolatility:   decimal.NewFromFloat(1.5),
		CalmMaxCurvature:    decimal.NewFromFloat(2.5),
		StormDurationHours:  4,
		StormMinPower:       decimal.NewFromFloat(4),
		StormMaxPower:       decimal.NewFromFloat(9),
		TakeProfitDiff:      decimal.NewFromFloat(3),
		StopLossDiff:        decimal.NewFromFloat(3),
		CheckDirectionHours: 24,
		CheckDirectionDiff:  decimal.NewFromFloat(0),
	}
}

func maxCBSParams() *params.CBS {
	return &params.CBS{
		CalmDurationHours:   17,
		CalmMaxVolatility:   decimal.NewFromFloat(2.5),
		CalmMaxCurvature:    decimal.NewFromFloat(3.5),
		StormDurationHours:  6,
		StormMinPower:       decimal.NewFromFloat(5),
		StormMaxPower:       decimal.NewFromFloat(11),
		TakeProfitDiff:      decimal.NewFromFloat(3),
		StopLossDiff:        decimal.NewFromFloat(3),
		CheckDirectionHours: 24,
		CheckDirectionDiff:  decimal.NewFromFloat(0),
	}
}

func minCBSRParams() *params.CBSR {
	return &params.CBSR{
		PeriodHoursMin:                 11,
		PeriodHoursMax:                 30,
		StormToCalmMin:                 decimal.NewFromFloat(0.24),
		StormToCalmMax:                 decimal.NewFromFloat(0.5),
		CalmMaxVolatilityToStormPower:  decimal.NewFromFloat(0.27),
		CalmMaxCurvatureToStormPower:   decimal.NewFromFloat(0.32),
		TakeProfitDiffToStormPower:     decimal.NewFromFloat(0.43),
		StopLossDiffToStormPower:       decimal.NewFromFloat(0.43),
		CalmToCheckDirection:           decimal.NewFromFloat(0.71),
		StormPowerToCheckDirectionDiff: decimal.NewFromFloat(5),
	}
}

func maxCBSRParams() *params.CBSR {
	return &params.CBSR{
		PeriodHoursMin:                 11,
		PeriodHoursMax:                 30,
		StormToCalmMin:                 decimal.NewFromFloat(0.28),
		StormToCalmMax:                 decimal.NewFromFloat(0.7),
		CalmMaxVolatilityToStormPower:  decimal.NewFromFloat(0.27),
		CalmMaxCurvatureToStormPower:   decimal.NewFromFloat(0.32),
		TakeProfitDiffToStormPower:     decimal.NewFromFloat(0.43),
		StopLossDiffToStormPower:       decimal.NewFromFloat(0.43),
		CalmToCheckDirection:           decimal.NewFromFloat(0.71),
		StormPowerToCheckDirectionDiff: decimal.NewFromFloat(5),
	}
}

func minFTParams() *params.FT {
	return &params.FT{
		TrendDurationHours:  21,
		TrendMaxVolatility:  decimal.NewFromFloat(3),
		TrendMinCurvature:   decimal.NewFromFloat(6),
		TrendMaxCurvature:   decimal.NewFromFloat(9),
		TakeProfitDiff:      decimal.NewFromFloat(4),
		StopLossDiff:        decimal.NewFromFloat(4),
		CheckDirectionHours: 24,
		CheckDirectionDiff:  decimal.NewFromFloat(0),
	}
}

func maxFTParams() *params.FT {
	return &params.FT{
		TrendDurationHours:  23,
		TrendMaxVolatility:  decimal.NewFromFloat(5),
		TrendMinCurvature:   decimal.NewFromFloat(9),
		TrendMaxCurvature:   decimal.NewFromFloat(12),
		TakeProfitDiff:      decimal.NewFromFloat(4),
		StopLossDiff:        decimal.NewFromFloat(4),
		CheckDirectionHours: 24,
		CheckDirectionDiff:  decimal.NewFromFloat(0),
	}
}

func (r optimizerApp) OptimizeParams(ctx context.Context, name, quoteSymbol string, from, to time.Time, minFrequency float64) error {
	modifyingParams := make([]decimal.Decimal, len(r.startParams))
	copy(modifyingParams, r.startParams)

	hours, err := r.candlestickRepository.GetCandlesticks(ctx, quoteSymbol, candlestick.IntervalHour, from, to)
	if err != nil {
		return err
	}

	modifier := params.NewBruteForceParamsModifier(
		r.minParams,
		r.maxParams,
		decimal.NewFromFloat(r.modifyRate),
	)

	var currentStats, bestStats paramsStats
	var frequentEnoughStats []paramsStats
	bar := pb.StartNew(modifier.GetTotalSteps())
	for modifier.Modify(modifyingParams) {
		count := 0
		accurate := 0
		total := 0
		countChan := make(chan bool)
		accurateChan := make(chan bool)
		totalChan := make(chan bool)
		var wg sync.WaitGroup
		for i := range hours {
			wg.Add(1)
			go r.calcStats(
				ctx,
				&wg,
				modifyingParams,
				hours[i],
				quoteSymbol,
				countChan,
				accurateChan,
				totalChan,
			)
			if (i > 0 && i%optimizerThreads == 0) || i == len(hours)-1 {
			outer:
				for {
					select {
					case <-countChan:
						count++
					case <-accurateChan:
						accurate++
					case <-totalChan:
						total++
					default:
						if total == i+1 {
							break outer
						}
					}
				}
				wg.Wait()
			}
		}
		frequency := float64(count) / float64(total) * 100
		accuracy := float64(accurate) / float64(count) * 100
		if frequency >= minFrequency {
			currentStats = paramsStats{
				params:    make([]decimal.Decimal, len(modifyingParams)),
				frequency: frequency,
				accuracy:  accuracy,
			}
			copy(currentStats.params, modifyingParams)
			if bestStats.accuracy < currentStats.accuracy {
				bestStats = currentStats
			} else if bestStats.accuracy == currentStats.accuracy && bestStats.frequency < currentStats.frequency {
				bestStats = currentStats
			}
			frequentEnoughStats = append(frequentEnoughStats, currentStats)
		}
		bar.Increment()
	}
	bar.Finish()

	return r.saveResults(name, bestStats, frequentEnoughStats)
}

func (r optimizerApp) saveResults(name string, bestStats paramsStats, frequentEnoughStats []paramsStats) error {
	fmt.Println("FREQUENT ENOUGH:")
	for i := range frequentEnoughStats {
		frequencyStr := strconv.FormatFloat(frequentEnoughStats[i].frequency, 'f', 2, 64)
		accuracyStr := strconv.FormatFloat(frequentEnoughStats[i].accuracy, 'f', 2, 64)
		fmt.Println(frequentEnoughStats[i].params, frequencyStr, accuracyStr)
	}
	if len(frequentEnoughStats) == 0 {
		fmt.Println("none")
	}

	fmt.Println("SAVED:")
	if bestStats.frequency > 0 {
		frequencyStr := strconv.FormatFloat(bestStats.frequency, 'f', 2, 64)
		accuracyStr := strconv.FormatFloat(bestStats.accuracy, 'f', 2, 64)
		if err := r.paramsRepository.SaveParams(
			name+
				"_"+
				frequencyStr+
				"_"+
				accuracyStr,
			bestStats.params,
		); err != nil {
			return err
		}
		fmt.Println(bestStats.params, frequencyStr, accuracyStr)
	} else {
		fmt.Println("none")
	}

	return nil
}

func (r optimizerApp) calcStats(
	ctx context.Context,
	wg *sync.WaitGroup,
	params []decimal.Decimal,
	hour candlestick.Candlestick,
	quoteSymbol string,
	countChan chan bool,
	accurateChan chan bool,
	totalChan chan bool,
) {
	defer wg.Done()
	a, _, err := r.adviser.GetAdvice(ctx, params, hour, quoteSymbol)
	if err != nil {
		panic(err)
	}

	if a != nil {
		countChan <- true
		expirationPeriod, err := r.candlestickRepository.GetCandlesticks(
			ctx,
			quoteSymbol,
			candlestick.IntervalHour,
			hour.Timestamp.Add(time.Hour),
			a.Expiration,
		)
		if err != nil {
			panic(err)
		}

		result, _ := r.calc.CalculateOrderResult(
			hour.Close,
			hour.Close.Sub(a.TakeProfit).Abs(),
			hour.Close.Sub(a.StopLoss).Abs(),
			expirationPeriod,
		)
		if (a.TakeProfit.GreaterThan(hour.Close) && result == candlestick.OrderResultProfitBuy) ||
			(a.TakeProfit.LessThan(hour.Close) && result == candlestick.OrderResultProfitSell) {
			accurateChan <- true
		}
	}
	totalChan <- true
}
