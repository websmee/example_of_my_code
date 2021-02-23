package candlestick

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

const (
	tradingStartHour = 1
	tradingEndHour   = 21
)

type basicFilter struct {
	repository Repository
}

func NewBasicFilter(repository Repository) Repository {
	return &basicFilter{repository}
}

func (r basicFilter) GetCandlesticks(ctx context.Context, symbol string, interval Interval, from, to time.Time) ([]Candlestick, error) {
	cs, err := r.repository.GetCandlesticks(ctx, symbol, interval, from, to)
	if err != nil {
		return nil, err
	}

	result := make([]Candlestick, len(cs))
	resultIndex := 0
	for i := range cs {
		if cs[i].Close.Equals(decimal.NewFromInt(0)) {
			continue
		}
		if interval == IntervalHour &&
			(cs[i].Timestamp.Hour() < tradingStartHour || cs[i].Timestamp.Hour() > tradingEndHour) {
			continue
		}
		result[resultIndex] = cs[i]
		resultIndex++
	}

	return result[:resultIndex], nil
}