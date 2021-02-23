package app

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/websmee/example_of_my_code/adviser/domain/advice"
	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
	"github.com/websmee/example_of_my_code/adviser/domain/params"
	"github.com/websmee/example_of_my_code/adviser/domain/signal"
)

type SignalTesterApp interface {
	TestSignal(ctx context.Context, name, quoteSymbol string, from, to time.Time) error
}

type cbsTesterApp struct {
	candlestickRepository candlestick.Repository
	paramsRepository      params.Repository
	adviser               advice.Adviser
	calc                  candlestick.Calculator
}

func NewCBSTesterApp(
	candlestickRepository candlestick.Repository,
	paramsRepository params.Repository,
) SignalTesterApp {
	calc := candlestick.DefaultCalculator()
	candlestickRepository = candlestick.NewBasicFilter(candlestickRepository)
	return &cbsTesterApp{
		candlestickRepository: candlestickRepository,
		paramsRepository:      paramsRepository,
		adviser:               advice.NewCBSAdviser(candlestickRepository, calc),
		calc:                  calc,
	}
}

func (r cbsTesterApp) TestSignal(ctx context.Context, name, quoteSymbol string, from, to time.Time) error {
	p, err := r.paramsRepository.LoadParams(name)
	if err != nil {
		return err
	}
	signalParams := new(signal.CBSParams)
	signalParams.SetParams(p)

	hours, err := r.candlestickRepository.GetCandlesticks(ctx, quoteSymbol, candlestick.IntervalHour, from, to)
	if err != nil {
		return err
	}

	var (
		CBSOk,
		CBSTooShortCalmPeriod,
		CBSNotCalmEnough,
		CBSTooSteepCalmLine,
		CBSTooWeakStorm,
		CBSTooStrongStorm,
		CBSSystemError int
	)
	count := 0
	accurate := 0
	for i := range hours {
		a, reason, err := r.adviser.GetAdvice(ctx, signalParams.GetParams(), hours[i], quoteSymbol)
		if err != nil {
			return err
		}

		switch reason {
		case advice.CBSOk:
			CBSOk++
		case advice.CBSTooShortCalmPeriod:
			CBSTooShortCalmPeriod++
		case advice.CBSNotCalmEnough:
			CBSNotCalmEnough++
		case advice.CBSTooSteepCalmLine:
			CBSTooSteepCalmLine++
		case advice.CBSTooWeakStorm:
			CBSTooWeakStorm++
		case advice.CBSTooStrongStorm:
			CBSTooStrongStorm++
		case advice.CBSSystemError:
			CBSSystemError++
		}

		if a != nil {
			count++
			expirationPeriod, err := r.candlestickRepository.GetCandlesticks(
				ctx,
				quoteSymbol,
				candlestick.IntervalHour,
				hours[i].Timestamp.Add(time.Hour),
				a.Expiration,
			)
			if err != nil {
				return err
			}

			result, _ := r.calc.CalculateOrderResult(
				hours[i].Close,
				hours[i].Close.Sub(a.TakeProfit).Abs(),
				hours[i].Close.Sub(a.StopLoss).Abs(),
				expirationPeriod,
			)
			if result == candlestick.OrderResultProfitBuy || result == candlestick.OrderResultProfitSell {
				accurate++
				fmt.Println(hours[i].Timestamp, a)
			}
		}
	}

	frequency := strconv.FormatFloat(float64(count)/float64(len(hours))*100, 'f', 2, 64)
	accuracy := strconv.FormatFloat(float64(accurate)/float64(count)*100, 'f', 2, 64)
	fmt.Println("Total", len(hours))
	fmt.Println("CBSOk", CBSOk)
	fmt.Println("CBSTooShortCalmPeriod", CBSTooShortCalmPeriod)
	fmt.Println("CBSNotCalmEnough", CBSNotCalmEnough)
	fmt.Println("CBSTooSteepCalmLine", CBSTooSteepCalmLine)
	fmt.Println("CBSTooWeakStorm", CBSTooWeakStorm)
	fmt.Println("CBSTooStrongStorm", CBSTooStrongStorm)
	fmt.Println("CBSSystemError", CBSSystemError)
	fmt.Println("FREQUENCY", frequency)
	fmt.Println("ACCURACY", accuracy)

	return nil
}
