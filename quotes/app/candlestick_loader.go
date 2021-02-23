package app

import (
	"fmt"
	"time"

	"github.com/websmee/example_of_my_code/quotes/domain/candlestick"
	"github.com/websmee/example_of_my_code/quotes/domain/quote"
)

const (
	startDate = "2020-01-01T00:00:00Z"
	endDate   = "2021-02-01T00:00:00Z"
)

type CandlestickLoader interface {
	LoadCandlesticks() error
}

type candlestickLoader struct {
	loader          candlestick.Loader
	candlestickRepo candlestick.Repository
	quoteRepo       quote.Repository
}

func NewCandlestickLoader(loader candlestick.Loader, candlestickRepo candlestick.Repository, quoteRepo quote.Repository) *candlestickLoader {
	return &candlestickLoader{
		loader:          loader,
		candlestickRepo: candlestickRepo,
		quoteRepo:       quoteRepo,
	}
}

func (r *candlestickLoader) LoadCandlesticks() error {
	quotes, err := r.quoteRepo.GetQuotes()
	if err != nil {
		return err
	}

	for _, q := range quotes {
		if err := r.load(&q, startDate, endDate, time.Hour*24, time.Millisecond*100, candlestick.IntervalHour); err != nil {
			return err
		}
		if err := r.load(&q, startDate, endDate, time.Hour*24*30, time.Millisecond*100, candlestick.IntervalDay); err != nil {
			return err
		}
		if err := r.load(&q, startDate, endDate, time.Hour*24*100, time.Millisecond*100, candlestick.IntervalMonth); err != nil {
			return err
		}
	}

	return nil
}

func (r *candlestickLoader) load(quote *quote.Quote, startDate string, endDate string, step time.Duration, delay time.Duration, interval candlestick.Interval) error {
	start, _ := time.Parse(time.RFC3339, startDate)
	end, _ := time.Parse(time.RFC3339, endDate)
	start = start.UTC()
	end = end.UTC()

	fmt.Printf("loading candlesticks %s\n", interval)

	for end.Sub(start) > 0 {
		stepEnd := start.Add(step)
		if end.Sub(stepEnd) < 0 {
			stepEnd = end
		}

		cs, err := r.loader.LoadCandlesticks(quote, start, stepEnd, interval)
		if err != nil {
			return err
		}

		for i := range cs {
			err := r.candlestickRepo.SaveCandlestick(&cs[i])
			if err != nil {
				return err
			}
		}

		fmt.Printf("candlesticks saved: %d\n", len(cs))

		start = stepEnd
		time.Sleep(delay)
	}

	fmt.Printf("loading candlesticks finished\n\n")

	return nil
}
