package params

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/advice"
	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
	"github.com/websmee/example_of_my_code/adviser/domain/quote"
)

const TestOrderExpirationPeriod = 30 * 24 * time.Hour

type AdviserParamsTester interface {
	TestParams(
		ctx context.Context,
		adviser advice.Adviser,
		prms []decimal.Decimal,
		quote quote.Quote,
		from, to time.Time,
		advicesChan chan []advice.InternalAdvice,
	)
	GetTotalSteps(ctx context.Context, quote quote.Quote, from, to time.Time) (int, error)
}

type adviserParamsTester struct {
	candlestickRepository candlestick.Repository
	calc                  candlestick.Calculator
}

func NewAdviserParamsTester(candlestickRepository candlestick.Repository) AdviserParamsTester {
	return &adviserParamsTester{
		candlestickRepository: candlestickRepository,
		calc:                  candlestick.NewDefaultCalculator(),
	}
}

func (r adviserParamsTester) TestParams(
	ctx context.Context,
	adviser advice.Adviser,
	prms []decimal.Decimal,
	quote quote.Quote,
	from, to time.Time,
	advicesChan chan []advice.InternalAdvice,
) {
	hours, err := r.candlestickRepository.GetCandlesticks(ctx, quote.Symbol, candlestick.IntervalHour, from, to)
	if err != nil {
		panic(err)
	}

	for i := range hours {
		advices, err := adviser.GetAdvices(ctx, prms, hours[i], quote.Symbol)
		if err != nil {
			panic(err)
		}

		for j := range advices {
			if advices[j].Status == advice.StatusOK {
				expirationPeriod, err := r.candlestickRepository.GetCandlesticks(
					ctx,
					quote.Symbol,
					candlestick.IntervalHour,
					hours[i].Timestamp.Add(time.Hour),
					hours[i].Timestamp.Add(TestOrderExpirationPeriod),
				)
				if err != nil {
					panic(err)
				}

				advices[j].OrderResult, advices[j].OrderClosed = r.calc.CalculateOrderResult(
					advices[j].CurrentPrice,
					advices[j].TakeProfit,
					advices[j].StopLoss,
					expirationPeriod,
				)
			}
		}

		advicesChan <- advices
	}
}

func (r adviserParamsTester) GetTotalSteps(ctx context.Context, quote quote.Quote, from, to time.Time) (int, error) {
	hours, err := r.candlestickRepository.GetCandlesticks(ctx, quote.Symbol, candlestick.IntervalHour, from, to)
	if err != nil {
		return 0, err
	}

	return len(hours), nil
}
