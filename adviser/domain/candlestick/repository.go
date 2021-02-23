package candlestick

import (
	"context"
	"time"
)

type Repository interface {
	GetCandlesticks(ctx context.Context, symbol string, interval Interval, from, to time.Time) ([]Candlestick, error)
}
