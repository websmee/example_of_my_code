package candlestick

import (
	"time"

	"github.com/websmee/example_of_my_code/quotes/domain/quote"
)

type Repository interface {
	SaveCandlestick(candlestick *Candlestick) error
	GetCandlesticks(quote *quote.Quote, interval Interval, from, to time.Time) ([]Candlestick, error)
	GetLastCandlestickTimestamp(quote *quote.Quote, interval Interval) (time.Time, error)
}
