package candlestick

import (
	"context"
	"time"
)

type GetterDirection string

const (
	GetterDirectionForward  GetterDirection = "forward"
	GetterDirectionBackward GetterDirection = "backward"
)

type GreedyGetter interface {
	GetCandlesticksByCount(
		ctx context.Context,
		symbol string,
		interval Interval,
		start time.Time,
		direction GetterDirection,
		count int,
	) ([]Candlestick, error)
}

type greedyGetter struct {
	candlestickRepository Repository
}

func NewGreedyGetter(
	candlestickRepository Repository,
) GreedyGetter {
	return &greedyGetter{
		candlestickRepository: candlestickRepository,
	}
}

func (r greedyGetter) GetCandlesticksByCount(
	ctx context.Context,
	symbol string,
	interval Interval,
	start time.Time,
	direction GetterDirection,
	count int,
) ([]Candlestick, error) {
	from := start
	to := start
	switch direction {
	case GetterDirectionForward:
		to = to.Add(time.Duration(count*10) * time.Hour)
	case GetterDirectionBackward:
		from = from.Add(-time.Duration(count*10) * time.Hour)
	}

	candlesticks, err := r.candlestickRepository.GetCandlesticks(ctx, symbol, interval, from, to)
	if err != nil {
		return nil, err
	}
	if len(candlesticks) <= count {
		return candlesticks, nil
	}

	result := make([]Candlestick, count)
	switch direction {
	case GetterDirectionForward:
		copy(result, candlesticks)
	case GetterDirectionBackward:
		copy(result, candlesticks[len(candlesticks)-count:])
	}

	return result, nil
}
