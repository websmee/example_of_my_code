package infrastructure

import (
	"time"

	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/chart"
	"github.com/piquette/finance-go/datetime"
	"github.com/pkg/errors"

	"github.com/websmee/example_of_my_code/quotes/domain/candlestick"
	"github.com/websmee/example_of_my_code/quotes/domain/quote"
)

type candlestickLoader struct {
}

func NewCandlestickYahooLoader() *candlestickLoader {
	return &candlestickLoader{}
}

func (r *candlestickLoader) LoadCandlesticks(quote *quote.Quote, start time.Time, end time.Time, interval candlestick.Interval) ([]candlestick.Candlestick, error) {
	params := &chart.Params{
		Symbol:   quote.Symbol,
		Start:    datetime.New(&start),
		End:      datetime.New(&end),
		Interval: datetime.Interval(interval),
	}
	iter := chart.Get(params)

	var cs []candlestick.Candlestick
	for iter.Next() {
		c := chartBarToCandlestick(iter.Bar())
		c.QuoteId = quote.Id
		c.Interval = interval
		cs = append(cs, c)
	}

	return cs, errors.Wrap(iter.Err(), "LoadCandlesticks failed")
}

func chartBarToCandlestick(b *finance.ChartBar) candlestick.Candlestick {
	return candlestick.Candlestick{
		Open:      b.Open,
		Low:       b.Low,
		High:      b.High,
		Close:     b.Close,
		AdjClose:  b.AdjClose,
		Volume:    b.Volume,
		Timestamp: time.Unix(int64(b.Timestamp), 0).UTC(),
	}
}
