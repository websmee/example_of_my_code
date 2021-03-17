package app

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"

	"github.com/websmee/example_of_my_code/quotes/domain/candlestick"
	"github.com/websmee/example_of_my_code/quotes/domain/quote"
)

type QuotesApp interface {
	GetQuotes() ([]quote.Quote, error)
	GetCandlesticks(symbol string, interval candlestick.Interval, from, to time.Time) ([]candlestick.Candlestick, error)
	HealthCheck() bool
}

const healthCheckQuoteSymbol = "AAPL"

type quotesApp struct {
	logger          log.Logger
	counter         metrics.Counter
	quoteRepo       quote.Repository
	candlestickRepo candlestick.Repository
}

func NewQuotesApp(
	logger log.Logger,
	counter metrics.Counter,
	quoteRepo quote.Repository,
	candlestickRepo candlestick.Repository,
) QuotesApp {
	var svc QuotesApp
	{
		svc = &quotesApp{
			logger:          logger,
			counter:         counter,
			quoteRepo:       quoteRepo,
			candlestickRepo: candlestickRepo,
		}
		svc = LoggingMiddleware(logger)(svc)
		svc = InstrumentingMiddleware(counter)(svc)
	}
	return svc
}

func (r *quotesApp) GetQuotes() ([]quote.Quote, error) {
	return r.quoteRepo.GetQuotes()
}

func (r *quotesApp) GetCandlesticks(symbol string, interval candlestick.Interval, from, to time.Time) ([]candlestick.Candlestick, error) {
	q, err := r.quoteRepo.GetQuote(symbol)
	if err != nil {
		return nil, err
	}

	return r.candlestickRepo.GetCandlesticks(q, interval, from, to)
}

func (r *quotesApp) HealthCheck() bool {
	_, err := r.quoteRepo.GetQuote(healthCheckQuoteSymbol)
	return err == nil
}
