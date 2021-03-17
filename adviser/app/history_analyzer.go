package app

import (
	"context"
	"time"

	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
)

type HistoryAnalyzerApp interface {
	Analyze(ctx context.Context, quoteSymbol string, from, to time.Time) error
}

type analyzerApp struct {
	candlestickRepository candlestick.Repository
	calc                  candlestick.Calculator
}

func NewAnalyzerApp(
	candlestickRepository candlestick.Repository,
) HistoryAnalyzerApp {
	calc := candlestick.DefaultCalculator()
	candlestickRepository = candlestick.NewBasicFilter(candlestickRepository)
	return &analyzerApp{
		candlestickRepository: candlestickRepository,
		calc:                  calc,
	}
}

func (r analyzerApp) Analyze(ctx context.Context, quoteSymbol string, from, to time.Time) error {
	//hours, err := r.candlestickRepository.GetCandlesticks(ctx, quoteSymbol, candlestick.IntervalHour, from, to)
	//if err != nil {
	//	return err
	//}
	//
	//stats := make(map[int]map[decimal.Decimal]int)
	//for i := range hours {
	//
	//}

	return nil
}
