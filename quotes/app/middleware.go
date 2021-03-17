package app

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"

	"github.com/websmee/example_of_my_code/quotes/domain/candlestick"
	"github.com/websmee/example_of_my_code/quotes/domain/quote"
)

type Middleware func(service QuotesApp) QuotesApp

func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next QuotesApp) QuotesApp {
		return loggingMiddleware{logger, next}
	}
}

type loggingMiddleware struct {
	logger log.Logger
	next   QuotesApp
}

func (mw loggingMiddleware) GetQuotes() (quotes []quote.Quote, err error) {
	defer func() {
		_ = mw.logger.Log("method", "GetQuotes", "error", err)
	}()
	return mw.next.GetQuotes()
}

func (mw loggingMiddleware) GetCandlesticks(symbol string, interval candlestick.Interval, from, to time.Time) (candlesticks []candlestick.Candlestick, err error) {
	defer func() {
		_ = mw.logger.Log("method", "GetCandlesticks", "symbol", symbol, "interval", interval, "from", from, "to", to, "error", err)
	}()
	return mw.next.GetCandlesticks(symbol, interval, from, to)
}

func (mw loggingMiddleware) HealthCheck() bool {
	return mw.next.HealthCheck()
}

func InstrumentingMiddleware(counter metrics.Counter) Middleware {
	return func(next QuotesApp) QuotesApp {
		return instrumentingMiddleware{
			counter: counter,
			next:    next,
		}
	}
}

type instrumentingMiddleware struct {
	counter metrics.Counter
	next    QuotesApp
}

func (mw instrumentingMiddleware) GetQuotes() ([]quote.Quote, error) {
	v, err := mw.next.GetQuotes()
	mw.counter.Add(float64(len(v)))
	return v, err
}

func (mw instrumentingMiddleware) GetCandlesticks(symbol string, interval candlestick.Interval, from, to time.Time) ([]candlestick.Candlestick, error) {
	v, err := mw.next.GetCandlesticks(symbol, interval, from, to)
	mw.counter.Add(float64(len(v)))
	return v, err
}

func (mw instrumentingMiddleware) HealthCheck() bool {
	return mw.next.HealthCheck()
}
