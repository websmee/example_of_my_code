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
	"github.com/websmee/example_of_my_code/adviser/domain/quote"
)

type ParamsOptimizerApp interface {
	OptimizeParams(ctx context.Context, name string, from, to time.Time, minFrequency float64) error
}

type optimizerApp struct {
	quoteRepository       quote.Repository
	candlestickRepository candlestick.Repository
	paramsRepository      params.Repository
	tester                params.AdviserParamsTester
	adviser               advice.Adviser
	calc                  candlestick.Calculator
	adviceSelector        advice.Selector
	startParams           []decimal.Decimal
	minParams             []decimal.Decimal
	maxParams             []decimal.Decimal
	modifyRate            float64
}

func NewCBSScaledOptimizerApp(
	quoteRepository quote.Repository,
	candlestickRepository candlestick.Repository,
	paramsRepository params.Repository,
	modifyRate float64,
) ParamsOptimizerApp {
	return newOptimizerApp(
		quoteRepository,
		candlestickRepository,
		paramsRepository,
		advice.NewCBSScaledAdviser(candlestickRepository, candlestick.NewDefaultCalculator()),
		minCBSScaledParams().GetParams(),
		minCBSScaledParams().GetParams(),
		maxCBSScaledParams().GetParams(),
		modifyRate,
	)
}

func newOptimizerApp(
	quoteRepository quote.Repository,
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
		quoteRepository:       quoteRepository,
		candlestickRepository: candlestickRepository,
		paramsRepository:      paramsRepository,
		tester:                params.NewAdviserParamsTester(candlestickRepository),
		adviser:               adviser,
		calc:                  candlestick.NewDefaultCalculator(),
		adviceSelector:        advice.NewDefaultSelector(),
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

func minCBSScaledParams() *advice.CBSScaledParams {
	return &advice.CBSScaledParams{
		PeriodHoursMin:                 20,
		PeriodHoursMax:                 40,
		StormToCalmMin:                 decimal.NewFromFloat(0.2),
		StormToCalmMax:                 decimal.NewFromFloat(0.4),
		StormMinPowerToCalmMaxChange:   decimal.NewFromFloat(3),
		StormMaxPowerToCalmMaxChange:   decimal.NewFromFloat(8),
		StormMinVolumeToCalmVolume:     decimal.NewFromFloat(0.5),
		CalmMaxChangeToStormPower:      decimal.NewFromFloat(0.35),
		CalmMaxCurvatureToStormPower:   decimal.NewFromFloat(0.1),
		TakeProfitDiffToStormPower:     decimal.NewFromFloat(0.4),
		StopLossDiffToStormPower:       decimal.NewFromFloat(0.4),
		CalmToCheckDirection:           decimal.NewFromFloat(0.3),
		StormPowerToCheckDirectionDiff: decimal.NewFromFloat(1),
	}
}

func maxCBSScaledParams() *advice.CBSScaledParams {
	return &advice.CBSScaledParams{
		PeriodHoursMin:                 30,
		PeriodHoursMax:                 50,
		StormToCalmMin:                 decimal.NewFromFloat(0.2),
		StormToCalmMax:                 decimal.NewFromFloat(0.4),
		StormMinPowerToCalmMaxChange:   decimal.NewFromFloat(3),
		StormMaxPowerToCalmMaxChange:   decimal.NewFromFloat(8),
		StormMinVolumeToCalmVolume:     decimal.NewFromFloat(0.5),
		CalmMaxChangeToStormPower:      decimal.NewFromFloat(0.35),
		CalmMaxCurvatureToStormPower:   decimal.NewFromFloat(0.1),
		TakeProfitDiffToStormPower:     decimal.NewFromFloat(0.4),
		StopLossDiffToStormPower:       decimal.NewFromFloat(0.4),
		CalmToCheckDirection:           decimal.NewFromFloat(0.3),
		StormPowerToCheckDirectionDiff: decimal.NewFromFloat(1),
	}
}

func minFTParams() *advice.FTParams {
	return &advice.FTParams{
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

func maxFTParams() *advice.FTParams {
	return &advice.FTParams{
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

func (r optimizerApp) OptimizeParams(ctx context.Context, name string, from, to time.Time, minFrequency float64) error {
	modifyingParams := make([]decimal.Decimal, len(r.startParams))
	copy(modifyingParams, r.startParams)

	quotes, err := r.quoteRepository.GetQuotes(ctx)
	if err != nil {
		return err
	}

	testerTotalSteps, err := r.getTesterTotalSteps(ctx, quotes, from, to)
	if err != nil {
		return err
	}

	modifier := params.NewBruteForceParamsModifier(
		r.minParams,
		r.maxParams,
		decimal.NewFromFloat(r.modifyRate),
	)

	var globalCount int
	var currentStats, bestStats paramsStats
	var frequentEnoughStats []paramsStats
	bar := pb.StartNew(modifier.GetTotalSteps() * testerTotalSteps)
	for modifier.Modify(modifyingParams) {
		var count, advicesOK, accurate, loss, expired int
		var wg sync.WaitGroup
		advicesChan := make(chan []advice.InternalAdvice)
		for i := range quotes {
			q := quotes[i]
			wg.Add(1)
			go func() {
				r.tester.TestParams(ctx, r.adviser, modifyingParams, q, from, to, advicesChan)
				wg.Done()
			}()
		}

		go func() {
			wg.Wait()
			close(advicesChan)
			bar.SetCurrent(int64(globalCount))
		}()

		for a := range advicesChan {
			globalCount++
			count++
			if globalCount%(testerTotalSteps/100) == 0 {
				bar.SetCurrent(int64(globalCount))
			}

			var okAdvices []advice.InternalAdvice
			for i := range a {
				if a[i].Status == advice.StatusOK {
					okAdvices = append(okAdvices, a[i])
				}
			}

			if selectedAdvice := r.adviceSelector.SelectAdvice(okAdvices); selectedAdvice != nil {
				advicesOK++
				switch selectedAdvice.OrderResult {
				case candlestick.OrderResultTakeProfit:
					accurate++
				case candlestick.OrderResultStopLoss:
					loss++
				case candlestick.OrderResultExpired:
					expired++
				}
			}
		}

		frequency := float64(advicesOK) / float64(count) * 100
		accuracy := float64(accurate) / float64(advicesOK) * 100
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

func (r optimizerApp) getTesterTotalSteps(ctx context.Context, quotes []quote.Quote, from, to time.Time) (int, error) {
	total := 0
	for i := range quotes {
		t, err := r.tester.GetTotalSteps(ctx, quotes[i], from, to)
		if err != nil {
			return 0, err
		}
		total += t
	}

	return total, nil
}
