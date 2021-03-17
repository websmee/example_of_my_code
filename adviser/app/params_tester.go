package app

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/advice"
	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
	"github.com/websmee/example_of_my_code/adviser/domain/params"
	"github.com/websmee/example_of_my_code/adviser/domain/quote"
)

type ParamsTesterApp interface {
	TestParams(ctx context.Context, name string, from, to time.Time) error
}

type testerApp struct {
	quoteRepository       quote.Repository
	candlestickRepository candlestick.Repository
	paramsRepository      params.Repository
	adviser               advice.Adviser
	calc                  candlestick.Calculator
}

func NewCBSTesterApp(
	quoteRepository quote.Repository,
	candlestickRepository candlestick.Repository,
	paramsRepository params.Repository,
) ParamsTesterApp {
	return newTesterApp(
		quoteRepository,
		candlestickRepository,
		paramsRepository,
		advice.NewCBSAdviser(candlestickRepository, candlestick.DefaultCalculator()),
	)
}

func NewCBSRTesterApp(
	quoteRepository quote.Repository,
	candlestickRepository candlestick.Repository,
	paramsRepository params.Repository,
) ParamsTesterApp {
	return newTesterApp(
		quoteRepository,
		candlestickRepository,
		paramsRepository,
		advice.NewCBSRAdviser(candlestickRepository, candlestick.DefaultCalculator()),
	)
}

func newTesterApp(
	quoteRepository quote.Repository,
	candlestickRepository candlestick.Repository,
	paramsRepository params.Repository,
	adviser advice.Adviser,
) ParamsTesterApp {
	candlestickRepository = candlestick.NewBasicFilter(candlestickRepository)
	return &testerApp{
		quoteRepository:       quoteRepository,
		candlestickRepository: candlestickRepository,
		paramsRepository:      paramsRepository,
		adviser:               adviser,
		calc:                  candlestick.DefaultCalculator(),
	}
}

func (r testerApp) TestParams(ctx context.Context, name string, from, to time.Time) error {
	p, err := r.paramsRepository.LoadParams(name)
	if err != nil {
		return err
	}

	quotes, err := r.quoteRepository.GetQuotes(ctx)
	if err != nil {
		return err
	}

	total, err := r.getTotalHours(ctx, quotes, from, to)
	if err != nil {
		return err
	}
	bar := pb.StartNew(total)

	results := new(testResults)
	results.reasonsCounts = make(map[advice.NoAdviceReason]int64)
	resultsChan := make(chan *testResults)
	incrementChan := make(chan bool)
	for i := range quotes {
		go r.testForQuote(ctx, p, quotes[i], from, to, resultsChan, incrementChan)
	}

outer:
	for {
		select {
		case <-incrementChan:
			bar.Increment()
		case r := <-resultsChan:
			results.add(r)
			if bar.Current() == int64(total) {
				close(incrementChan)
				close(resultsChan)
				break outer
			}
		}
	}

	bar.Finish()

	frequency := strconv.FormatFloat(float64(results.count)/float64(total)*100, 'f', 2, 64)
	accuracy := strconv.FormatFloat(float64(results.accurate)/float64(results.count)*100, 'f', 2, 64)
	fmt.Println("Total", total)
	r.printReasons(results.reasonsCounts)
	fmt.Println("FREQUENCY", frequency)
	fmt.Println("ACCURACY", accuracy)
	fmt.Println("LOSS", results.loss)
	fmt.Println("EXPIRED", results.expired)

	return nil
}

type testResults struct {
	reasonsCounts map[advice.NoAdviceReason]int64
	count         int
	accurate      int
	loss          int
	expired       int
}

func (r *testResults) add(add *testResults) {
	for k := range add.reasonsCounts {
		r.reasonsCounts[k] += add.reasonsCounts[k]
	}
	r.count += add.count
	r.accurate += add.accurate
	r.loss += add.loss
	r.expired += add.expired
}

func (r testerApp) testForQuote(
	ctx context.Context,
	p []decimal.Decimal,
	quote quote.Quote,
	from, to time.Time,
	resultsChan chan *testResults,
	incrementChan chan bool,
) {
	hours, err := r.candlestickRepository.GetCandlesticks(ctx, quote.Symbol, candlestick.IntervalHour, from, to)
	if err != nil {
		panic(err)
	}

	results := new(testResults)
	results.reasonsCounts = make(map[advice.NoAdviceReason]int64)
	for j := range hours {
		a, reason, err := r.adviser.GetAdvice(ctx, p, hours[j], quote.Symbol)
		if err != nil {
			panic(err)
		}

		results.reasonsCounts[reason]++

		if a != nil {
			results.count++
			expirationPeriod, err := r.candlestickRepository.GetCandlesticks(
				ctx,
				quote.Symbol,
				candlestick.IntervalHour,
				hours[j].Timestamp.Add(time.Hour),
				a.Expiration,
			)
			if err != nil {
				panic(err)
			}

			result, _ := r.calc.CalculateOrderResult(
				hours[j].Close,
				hours[j].Close.Sub(a.TakeProfit).Abs(),
				hours[j].Close.Sub(a.StopLoss).Abs(),
				expirationPeriod,
			)
			switch {
			case (a.TakeProfit.GreaterThan(hours[j].Close) && result == candlestick.OrderResultProfitBuy) ||
				(a.TakeProfit.LessThan(hours[j].Close) && result == candlestick.OrderResultProfitSell):
				results.accurate++
			case result == candlestick.OrderResultExpired:
				results.expired++
			default:
				results.loss++
			}
		}
		incrementChan <- true
	}
	resultsChan <- results
}

func (r testerApp) getTotalHours(ctx context.Context, quotes []quote.Quote, from, to time.Time) (int, error) {
	total := 0
	for i := range quotes {
		hours, err := r.candlestickRepository.GetCandlesticks(ctx, quotes[i].Symbol, candlestick.IntervalHour, from, to)
		if err != nil {
			return 0, err
		}
		total += len(hours)
	}

	return total, nil
}

func (r testerApp) printReasons(reasonsCounts map[advice.NoAdviceReason]int64) {
	reasons := make([]int, len(reasonsCounts))
	i := 0
	for k := range reasonsCounts {
		reasons[i] = int(k)
		i++
	}
	sort.Ints(reasons)

	for _, k := range reasons {
		rn := advice.NoAdviceReason(k)
		fmt.Println(r.adviser.GetReasonName(rn), reasonsCounts[rn])
	}
}
