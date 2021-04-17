package infrastructure

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/quotes/domain/candlestick"
	"github.com/websmee/example_of_my_code/quotes/domain/quote"
	"github.com/websmee/example_of_my_code/quotes/infrastructure/tiingo"
)

type tiingoCandlestickLoader struct {
	client tiingo.Client
}

func NewTiingoCandlestickLoader(client tiingo.Client) candlestick.Loader {
	return &tiingoCandlestickLoader{
		client: client,
	}
}

func (r tiingoCandlestickLoader) LoadHistory(quote quote.Quote, start time.Time, end time.Time, interval candlestick.Interval) ([]candlestick.Candlestick, error) {
	prices, err := r.client.GetPrices(tiingo.PricesRequest{
		Ticker:       quote.Symbol,
		StartDate:    start,
		EndDate:      end,
		ResampleFreq: intervalToResampleFreq(interval),
	})
	if err != nil {
		return nil, err
	}

	candlesticks := make([]candlestick.Candlestick, len(prices))
	for i := range prices {
		candlesticks[i] = pricesToCandlestick(prices[i], quote.ID, interval)
	}

	return candlesticks, nil
}

func (r tiingoCandlestickLoader) LoadLatest(quote quote.Quote, interval candlestick.Interval) ([]candlestick.Candlestick, error) {
	return r.LoadHistory(quote, time.Now().Add(-24*time.Hour), time.Now(), interval)
}

func intervalToResampleFreq(_ candlestick.Interval) tiingo.ResponseResampleFreq {
	return tiingo.ResponseResampleFreqHour
}

func pricesToCandlestick(prices tiingo.Prices, quoteID int64, interval candlestick.Interval) candlestick.Candlestick {
	return candlestick.Candlestick{
		Open:      decimal.NewFromFloat(prices.Open),
		Low:       decimal.NewFromFloat(prices.Low),
		High:      decimal.NewFromFloat(prices.High),
		Close:     decimal.NewFromFloat(prices.Close),
		AdjClose:  decimal.NewFromInt(0),
		Volume:    int(prices.Volume),
		Timestamp: prices.Date,
		Interval:  interval,
		QuoteID:   quoteID,
	}
}
