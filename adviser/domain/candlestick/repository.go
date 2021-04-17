package candlestick

import (
	"context"
	"time"
)

type Repository interface {
	GetCandlesticks(ctx context.Context, symbol string, interval Interval, from, to time.Time) ([]Candlestick, error)
	GetCandlesticksByCount(ctx context.Context, symbol string, interval Interval, start time.Time, direction GetterDirection, count int) ([]Candlestick, error)
}
