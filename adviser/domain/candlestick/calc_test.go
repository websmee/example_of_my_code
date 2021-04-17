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

	calc := NewDefaultCalculator()
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

	calc := NewDefaultCalculator()

	{
		// BUY TAKE PROFIT
		currentPrice := decimal.NewFromInt(3)
		takeProfitPrice := decimal.NewFromInt(7)
		stopLossPrice := decimal.NewFromInt(1)
		expirationPeriod := candlesticks
		r, _ := calc.CalculateOrderResult(currentPrice, takeProfitPrice, stopLossPrice, expirationPeriod)
		if r != OrderResultTakeProfit {
			t.Error(r)
		}
	}

	{
		// SELL TAKE PROFIT
		currentPrice := decimal.NewFromInt(6)
		takeProfitPrice := decimal.NewFromInt(5)
		stopLossPrice := decimal.NewFromInt(7)
		expirationPeriod := candlesticks
		r, _ := calc.CalculateOrderResult(currentPrice, takeProfitPrice, stopLossPrice, expirationPeriod)
		if r != OrderResultTakeProfit {
			t.Error(r)
		}
	}

	{
		// EXPIRED
		currentPrice := decimal.NewFromInt(6)
		takeProfitPrice := decimal.NewFromInt(11)
		stopLossPrice := decimal.NewFromInt(4)
		expirationPeriod := candlesticks[:1]
		r, _ := calc.CalculateOrderResult(currentPrice, takeProfitPrice, stopLossPrice, expirationPeriod)
		if r != OrderResultExpired {
			t.Error(r)
		}
	}

	{
		// STOP LOSS
		currentPrice := decimal.NewFromInt(8)
		takeProfitPrice := decimal.NewFromInt(14)
		stopLossPrice := decimal.NewFromInt(7)
		expirationPeriod := candlesticks
		r, _ := calc.CalculateOrderResult(currentPrice, takeProfitPrice, stopLossPrice, expirationPeriod)
		if r != OrderResultStopLoss {
			t.Error(r)
		}
	}
}
