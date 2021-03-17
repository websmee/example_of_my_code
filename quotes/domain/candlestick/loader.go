package candlestick

import (
	"time"

	"github.com/websmee/example_of_my_code/quotes/domain/quote"
)

type Loader interface {
	LoadCandlesticks(quote quote.Quote, start, end time.Time, interval Interval) ([]Candlestick, error)
}
