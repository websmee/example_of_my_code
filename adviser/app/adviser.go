package app

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/advice"
	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
	"github.com/websmee/example_of_my_code/adviser/domain/params"
	"github.com/websmee/example_of_my_code/adviser/domain/quote"
)

type AdviserApp interface {
	GetAdvices(ctx context.Context) ([]advice.Advice, error)
	HealthCheck() bool
}

type adviserApp struct {
	counter               metrics.Counter
	quoteRepository       quote.Repository
	candlestickRepository candlestick.Repository
	advisers              map[advice.AdviserType]advice.Adviser
	paramsRepository      params.Repository
}

func NewAdviserApp(
	logger log.Logger,
	counter metrics.Counter,
	quoteRepository quote.Repository,
	candlestickRepository candlestick.Repository,
	paramsRepository params.Repository,
) AdviserApp {
	var svc AdviserApp
	{
		svc = &adviserApp{
			counter:               counter,
			quoteRepository:       quoteRepository,
			candlestickRepository: candlestickRepository,
			advisers: map[advice.AdviserType]advice.Adviser{
				advice.AdviserTypeCBS: advice.NewCBSScaledAdviser(candlestickRepository, candlestick.NewDefaultCalculator()),
			},
			paramsRepository: paramsRepository,
		}
		svc = AdviserLoggingMiddleware(logger)(svc)
		svc = AdviserInstrumentingMiddleware(counter)(svc)
	}
	return svc
}

func (r adviserApp) GetAdvices(ctx context.Context) ([]advice.Advice, error) {
	var advices []advice.Advice
	for t, adviser := range r.advisers {
		adviserParams, err := r.paramsRepository.LoadParams(string(t))
		if err != nil {
			return nil, err
		}

		quotes, err := r.quoteRepository.GetQuotes(ctx)
		if err != nil {
			return nil, err
		}

		for i := range quotes {
			current, err := r.candlestickRepository.GetCandlesticks(
				ctx, quotes[i].Symbol,
				candlestick.IntervalHour,
				time.Now().Add(-2*time.Hour),
				time.Now().Add(-time.Hour),
			)
			if err != nil {
				return nil, err
			}
			if len(current) != 1 {
				continue
			}

			a, err := adviser.GetAdvices(ctx, adviserParams, current[0], quotes[i].Symbol)
			if err != nil {
				return nil, err
			}
			advices = append(advices, r.convertAdvices(a, quotes[i])...)
		}
	}

	return advices, nil
}

func (r adviserApp) convertAdvices(internal []advice.InternalAdvice, quote quote.Quote) []advice.Advice {
	advices := make([]advice.Advice, len(internal))
	for i := range internal {
		advices[i] = advice.Advice{
			Quote:            quote,
			Candlesticks:     nil, //todo
			Price:            internal[i].CurrentPrice,
			Amount:           decimal.NewFromInt(0), //todo
			TakeProfitPrice:  internal[i].TakeProfit,
			TakeProfitAmount: decimal.NewFromInt(0), //todo
			StopLossPrice:    internal[i].StopLoss,
			StopLossAmount:   decimal.NewFromInt(0), //todo
			Leverage:         0,                     //todo
			ExpiresAt:        time.Time{},           //todo
		}
	}

	return advices
}

func (r adviserApp) HealthCheck() bool {
	//todo
	return true
}
