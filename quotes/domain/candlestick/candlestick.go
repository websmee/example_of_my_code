package candlestick

import (
	"time"

	"github.com/shopspring/decimal"
)

type Interval string

const (
	IntervalHour Interval = "1h"
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
	QuoteID   int64
}
