package app

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"

	"github.com/websmee/example_of_my_code/quotes/domain/candlestick"
	"github.com/websmee/example_of_my_code/quotes/domain/quote"
)

type QuotesMiddleware func(service QuotesApp) QuotesApp

func QuotesLoggingMiddleware(logger log.Logger) QuotesMiddleware {
	return func(next QuotesApp) QuotesApp {
		return quotesLoggingMiddleware{logger, next}
	}
}

type quotesLoggingMiddleware struct {
	logger log.Logger
	next   QuotesApp
}

func (mw quotesLoggingMiddleware) GetQuotes() (quotes []quote.Quote, err error) {
	defer func() {
		_ = mw.logger.Log("method", "GetQuotes", "error", err)
	}()
	return mw.next.GetQuotes()
}

func (mw quotesLoggingMiddleware) GetCandlesticks(symbol string, interval candlestick.Interval, from, to time.Time) (candlesticks []candlestick.Candlestick, err error) {
	defer func() {
		_ = mw.logger.Log("method", "GetCandlesticks", "symbol", symbol, "interval", interval, "from", from, "to", to, "error", err)
	}()
	return mw.next.GetCandlesticks(symbol, interval, from, to)
}

func (mw quotesLoggingMiddleware) HealthCheck() bool {
	return mw.next.HealthCheck()
}

func QuotesInstrumentingMiddleware(counter metrics.Counter) QuotesMiddleware {
	return func(next QuotesApp) QuotesApp {
		return quotesInstrumentingMiddleware{
			counter: counter,
			next:    next,
		}
	}
}

type quotesInstrumentingMiddleware struct {
	counter metrics.Counter
	next    QuotesApp
}

func (mw quotesInstrumentingMiddleware) GetQuotes() ([]quote.Quote, error) {
	v, err := mw.next.GetQuotes()
	mw.counter.Add(float64(len(v)))
	return v, err
}

func (mw quotesInstrumentingMiddleware) GetCandlesticks(symbol string, interval candlestick.Interval, from, to time.Time) ([]candlestick.Candlestick, error) {
	v, err := mw.next.GetCandlesticks(symbol, interval, from, to)
	mw.counter.Add(float64(len(v)))
	return v, err
}

func (mw quotesInstrumentingMiddleware) HealthCheck() bool {
	return mw.next.HealthCheck()
}
