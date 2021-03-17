package dataset

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/advice"
	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
	"github.com/websmee/example_of_my_code/adviser/domain/nn"
	"github.com/websmee/example_of_my_code/adviser/domain/params"
)

type Factory interface {
	CreateCBSDataset(ctx context.Context, cbsParams *params.CBS, quoteSymbol string, from, to time.Time) (*Dataset, error)
}

type factory struct {
	candlestickRepository candlestick.Repository
}

func NewFactory(
	candlestickRepository candlestick.Repository,
) Factory {
	return &factory{
		candlestickRepository: candlestickRepository,
	}
}

func (r factory) CreateCBSDataset(ctx context.Context, cbsParams *params.CBS, quoteSymbol string, from, to time.Time) (*Dataset, error) {
	hours, err := r.candlestickRepository.GetCandlesticks(ctx, quoteSymbol, candlestick.IntervalHour, from, to)
	if err != nil {
		return nil, err
	}

	ds := NewCBSDataset()
	ds.Rows = make([]*Row, 0, len(hours))

	calc := candlestick.DefaultCalculator()
	adviser := advice.NewCBSAdviser(r.candlestickRepository, calc)

	for i := range hours {
		a, _, err := adviser.GetAdvice(ctx, cbsParams.GetParams(), hours[i], quoteSymbol)
		if err != nil {
			return nil, err
		}
		if a == nil {
			continue
		}

		expirationPeriod, err := r.candlestickRepository.GetCandlesticks(
			ctx,
			quoteSymbol,
			candlestick.IntervalHour,
			hours[i].Timestamp.Add(time.Hour),
			a.Expiration,
		)
		if err != nil {
			return nil, err
		}

		result, _ := calc.CalculateOrderResult(
			hours[i].Close,
			hours[i].Close.Sub(a.TakeProfit),
			hours[i].Close.Sub(a.StopLoss),
			expirationPeriod,
		)

		profitBuyProbability := decimal.NewFromFloat(0.01)
		profitSellProbability := decimal.NewFromFloat(0.01)
		stopLossProbability := decimal.NewFromFloat(0.01)
		switch result {
		case candlestick.OrderResultProfitBuy:
			profitBuyProbability = decimal.NewFromFloat(0.99)
		case candlestick.OrderResultProfitSell:
			profitSellProbability = decimal.NewFromFloat(0.99)
		case candlestick.OrderResultStopLoss:
			stopLossProbability = decimal.NewFromFloat(0.99)
		case candlestick.OrderResultExpired:
			stopLossProbability = decimal.NewFromFloat(0.99)
		}

		ds.Rows = append(
			ds.Rows,
			ds.NewRow().
				NewInput(decimal.NewFromFloat(123.45)). // todo: news data goes here
				NewOutput(profitBuyProbability).
				NewOutput(profitSellProbability).
				NewOutput(stopLossProbability).
				NewDebug(hours[i].Close).
				NewDebug(hours[i].Close.Sub(a.TakeProfit)).
				NewDebug(hours[i].Close.Sub(a.StopLoss)).
				NewDebug(decimal.NewFromInt(hours[i].Timestamp.Unix())),
		)

		if len(ds.Rows[len(ds.Rows)-1].OriginalInputs) != nn.CBSInputs {
			panic("invalid dataset row input size")
		}

		if len(ds.Rows[len(ds.Rows)-1].Outputs) != nn.CBSOutputs {
			panic("invalid dataset row output size")
		}
	}

	return ds, nil
}
