package candlestick

import (
	"time"

	"github.com/shopspring/decimal"
)

type Interval string

const (
	IntervalMinute Interval = "1m"
	IntervalHour   Interval = "1h"
	IntervalDay    Interval = "1d"
	IntervalMonth  Interval = "1mo"
)

type Candlestick struct {
	Open      decimal.Decimal `pg:",use_zero"`
	Low       decimal.Decimal `pg:",use_zero"`
	High      decimal.Decimal `pg:",use_zero"`
	Close     decimal.Decimal `pg:",use_zero"`
	AdjClose  decimal.Decimal `pg:",use_zero"`
	Volume    int             `pg:",use_zero"`
	Timestamp time.Time
	Interval  Interval
	QuoteId   int64
}

func (r Interval) Minutes() int {
	switch r {
	case IntervalMinute:
		return 1
	case IntervalHour:
		return 60
	case IntervalDay:
		return 60 * 24
	case IntervalMonth:
		return 60 * 24 * 30
	}

	return 0
}
