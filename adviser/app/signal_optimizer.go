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
	"github.com/websmee/example_of_my_code/adviser/domain/signal"
)

const optimizerThreads = 16

type SignalOptimizerApp interface {
	OptimizeSignal(ctx context.Context, name, quoteSymbol string, from, to time.Time, minFrequency float64) error
}

type cbsOptimizerApp struct {
	candlestickRepository candlestick.Repository
	paramsRepository      params.Repository
	adviser               advice.Adviser
	modifier              signal.ParamsModifier
	calc                  candlestick.Calculator
}

func NewCBSOptimizerApp(
	candlestickRepository candlestick.Repository,
	paramsRepository params.Repository,
	modifyRate float64,
) SignalOptimizerApp {
	calc := candlestick.DefaultCalculator()
	candlestickRepository = candlestick.NewBasicFilter(candlestickRepository)
	return &cbsOptimizerApp{
		candlestickRepository: candlestickRepository,
		paramsRepository:      paramsRepository,
		adviser:               advice.NewCBSAdviser(candlestickRepository, calc),
		modifier: signal.NewBruteForceParamsModifier(
			minCBSParams().GetParams(),
			maxCBSParams().GetParams(),
			decimal.NewFromFloat(modifyRate),
		),
		calc: calc,
	}
}

type signalStats struct {
	params    []decimal.Decimal
	frequency float64
	accuracy  float64
}

func minCBSParams() *signal.CBSParams {
	return &signal.CBSParams{
		CalmDurationHours:  18,
		CalmMaxVolatility:  decimal.NewFromInt(4),
		CalmMaxCurvature:   decimal.NewFromInt(7),
		StormDurationHours: 2,
		StormMinPower:      decimal.NewFromInt(13),
		StormMaxPower:      decimal.NewFromInt(18),
		TakeProfitDiff:     decimal.NewFromInt(4),
		StopLossDiff:       decimal.NewFromInt(2),
	}
}

func maxCBSParams() *signal.CBSParams {
	return &signal.CBSParams{
		CalmDurationHours:  22,
		CalmMaxVolatility:  decimal.NewFromInt(6),
		CalmMaxCurvature:   decimal.NewFromInt(9),
		StormDurationHours: 4,
		StormMinPower:      decimal.NewFromInt(18),
		StormMaxPower:      decimal.NewFromInt(25),
		TakeProfitDiff:     decimal.NewFromInt(5),
		StopLossDiff:       decimal.NewFromInt(3),
	}
}

func (r cbsOptimizerApp) OptimizeSignal(ctx context.Context, name, quoteSymbol string, from, to time.Time, minFrequency float64) error {
	signalParams := minCBSParams().GetParams()
	hours, err := r.candlestickRepository.GetCandlesticks(ctx, quoteSymbol, candlestick.IntervalHour, from, to)
	if err != nil {
		return err
	}

	var currentStats, bestStats signalStats
	var frequentEnoughStats []signalStats
	bar := pb.StartNew(r.modifier.GetTotalSteps())
	for r.modifier.Modify(signalParams) {
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
				signalParams,
				hours[i],
				quoteSymbol,
				countChan,
				accurateChan,
				totalChan,
			)
			if i > 0 && i%optimizerThreads == 0 {
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
			currentStats = signalStats{
				params:    signalParams,
				frequency: frequency,
				accuracy:  accuracy,
			}
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

func (r cbsOptimizerApp) saveResults(name string, bestStats signalStats, frequentEnoughStats []signalStats) error {
	fmt.Println("FREQUENT ENOUGH:")
	for i := range frequentEnoughStats {
		frequencyStr := strconv.FormatFloat(frequentEnoughStats[i].frequency, 'f', 2, 64)
		accuracyStr := strconv.FormatFloat(frequentEnoughStats[i].accuracy, 'f', 2, 64)
		fmt.Println(frequentEnoughStats[i].params, frequencyStr, accuracyStr)
	}

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
	fmt.Println("SAVED:")
	fmt.Println(bestStats.params, frequencyStr, accuracyStr)

	return nil
}

func (r cbsOptimizerApp) calcStats(
	ctx context.Context,
	wg *sync.WaitGroup,
	signalParams []decimal.Decimal,
	hour candlestick.Candlestick,
	quoteSymbol string,
	countChan chan bool,
	accurateChan chan bool,
	totalChan chan bool,
) {
	defer wg.Done()
	a, _, err := r.adviser.GetAdvice(ctx, signalParams, hour, quoteSymbol)
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
		if result == candlestick.OrderResultProfitBuy || result == candlestick.OrderResultProfitSell {
			accurateChan <- true
		}
	}
	totalChan <- true
}
