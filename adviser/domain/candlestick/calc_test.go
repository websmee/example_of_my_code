package candlestick

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalculator_CalculateVolatility(t *testing.T) {
	var candlesticks []Candlestick
	candlesticks = append(candlesticks, Candlestick{
		Open:  decimal.NewFromInt(5),
		Close: decimal.NewFromInt(5),
		High:  decimal.NewFromInt(5),
		Low:   decimal.NewFromInt(5),
	})
	candlesticks = append(candlesticks, Candlestick{
		Open:  decimal.NewFromInt(7),
		Close: decimal.NewFromInt(7),
		High:  decimal.NewFromInt(7),
		Low:   decimal.NewFromInt(7),
	})
	candlesticks = append(candlesticks, Candlestick{
		Open:  decimal.NewFromInt(12),
		Close: decimal.NewFromInt(12),
		High:  decimal.NewFromInt(12),
		Low:   decimal.NewFromInt(12),
	})

	calc := DefaultCalculator()
	volatility := calc.CalculateVolatility(candlesticks)
	expected := decimal.NewFromFloat(2.943920288775949)
	if !volatility.Equals(expected) {
		t.Error(volatility, expected)
	}
}

func TestCalculator_CalculateResults(t *testing.T) {
	var candlesticks []Candlestick
	candlesticks = append(candlesticks, Candlestick{
		Open:  decimal.NewFromInt(5),
		Close: decimal.NewFromInt(5),
		High:  decimal.NewFromInt(5),
		Low:   decimal.NewFromInt(5),
	})
	candlesticks = append(candlesticks, Candlestick{
		Open:  decimal.NewFromInt(7),
		Close: decimal.NewFromInt(7),
		High:  decimal.NewFromInt(7),
		Low:   decimal.NewFromInt(7),
	})
	candlesticks = append(candlesticks, Candlestick{
		Open:  decimal.NewFromInt(12),
		Close: decimal.NewFromInt(12),
		High:  decimal.NewFromInt(12),
		Low:   decimal.NewFromInt(12),
	})

	calc := DefaultCalculator()

	{
		// BUY TAKE PROFIT
		currentPrice := decimal.NewFromInt(3)
		takeProfitPriceDiff := decimal.NewFromInt(4)
		stopLossPriceDiff := takeProfitPriceDiff.Mul(decimal.NewFromFloat(0.3))
		expirationPeriod := candlesticks
		r, h := calc.CalculateOrderResult(currentPrice, takeProfitPriceDiff, stopLossPriceDiff, expirationPeriod)
		if r != OrderResultProfitBuy {
			t.Error(r)
		}
		if !h.Equals(decimal.NewFromInt(2)) {
			t.Error(h)
		}
	}

	{
		// SELL TAKE PROFIT
		currentPrice := decimal.NewFromInt(6)
		takeProfitPriceDiff := decimal.NewFromInt(1)
		stopLossPriceDiff := takeProfitPriceDiff.Mul(decimal.NewFromFloat(0.3))
		expirationPeriod := candlesticks
		r, h := calc.CalculateOrderResult(currentPrice, takeProfitPriceDiff, stopLossPriceDiff, expirationPeriod)
		if r != OrderResultProfitSell {
			t.Error(r)
		}
		if !h.Equals(decimal.NewFromInt(1)) {
			t.Error(h)
		}
	}

	{
		// EXPIRED
		currentPrice := decimal.NewFromInt(6)
		takeProfitPriceDiff := decimal.NewFromInt(5)
		stopLossPriceDiff := takeProfitPriceDiff.Mul(decimal.NewFromFloat(0.3))
		expirationPeriod := candlesticks[:1]
		r, h := calc.CalculateOrderResult(currentPrice, takeProfitPriceDiff, stopLossPriceDiff, expirationPeriod)
		if r != OrderResultExpired {
			t.Error(r)
		}
		if !h.Equals(decimal.NewFromInt(1)) {
			t.Error(h)
		}
	}

	{
		// STOP LOSS
		currentPrice := decimal.NewFromInt(8)
		takeProfitPriceDiff := decimal.NewFromInt(6)
		stopLossPriceDiff := takeProfitPriceDiff.Mul(decimal.NewFromFloat(0.3))
		expirationPeriod := candlesticks
		r, h := calc.CalculateOrderResult(currentPrice, takeProfitPriceDiff, stopLossPriceDiff, expirationPeriod)
		if r != OrderResultStopLoss {
			t.Error(r)
		}
		if !h.Equals(decimal.NewFromInt(3)) {
			t.Error(h)
		}
	}
}
