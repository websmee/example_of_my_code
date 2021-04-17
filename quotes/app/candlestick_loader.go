package app

import (
	"fmt"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/websmee/example_of_my_code/quotes/domain/candlestick"
	"github.com/websmee/example_of_my_code/quotes/domain/quote"
)

const (
	startDate = "2018-01-01T00:00:00Z"
	endDate   = "2021-05-01T00:00:00Z"
)

type CandlestickLoader interface {
	LoadCandlesticks() error
	LoadLatest() error
}

type candlestickLoader struct {
	loader          candlestick.Loader
	candlestickRepo candlestick.Repository
	quoteRepo       quote.Repository
}

func NewCandlestickLoader(
	logger log.Logger,
	loader candlestick.Loader,
	candlestickRepo candlestick.Repository,
	quoteRepo quote.Repository,
) CandlestickLoader {
	var svc CandlestickLoader
	{
		svc = &candlestickLoader{
			loader:          loader,
			candlestickRepo: candlestickRepo,
			quoteRepo:       quoteRepo,
		}
		svc = CandlestickLoaderLoggingMiddleware(logger)(svc)
	}
	return svc
}

func (r candlestickLoader) LoadCandlesticks() error {
	quotes, err := r.quoteRepo.GetQuotes(quote.StatusNew)
	if err != nil {
		return err
	}

	for _, q := range quotes {
		if err := r.load(q, startDate, endDate, candlestick.IntervalHour); err != nil {
			return err
		}
		fmt.Println("READY", q)
	}

	return nil
}

func (r candlestickLoader) LoadLatest() error {
	quotes, err := r.quoteRepo.GetQuotes(quote.StatusReady)
	if err != nil {
		return err
	}

	for _, q := range quotes {
		cs, err := r.loader.LoadLatest(q, candlestick.IntervalHour)
		if err != nil {
			return err
		}

		for i := range cs {
			err := r.candlestickRepo.SaveCandlestick(&cs[i])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r candlestickLoader) load(q quote.Quote, startDate string, endDate string, interval candlestick.Interval) error {
	start, _ := time.Parse(time.RFC3339, startDate)
	end, _ := time.Parse(time.RFC3339, endDate)
	start = start.UTC()
	end = end.UTC()

	cs, err := r.loader.LoadHistory(q, start, end, interval)
	if err != nil {
		return err
	}

	for i := range cs {
		err := r.candlestickRepo.SaveCandlestick(&cs[i])
		if err != nil {
			return err
		}
	}

	if err := r.quoteRepo.UpdateQuoteStatus(&q, quote.StatusReady); err != nil {
		return err
	}

	return nil
}
