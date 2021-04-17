package candlestick

import (
	"math"
	"time"

	"github.com/shopspring/decimal"
)

type OrderResult string

const (
	OrderResultNone       OrderResult = "none"
	OrderResultTakeProfit OrderResult = "profit"
	OrderResultStopLoss   OrderResult = "loss"
	OrderResultExpired    OrderResult = "expired"
)

type Calculator interface {
	CalculateMaxChange(candlesticks []Candlestick) decimal.Decimal
	CalculateSMA(candlesticks []Candlestick) decimal.Decimal
	CalculateVolatility(candlesticks []Candlestick) decimal.Decimal
	CalculateVolume(candlesticks []Candlestick) decimal.Decimal
	CalculateOrderResult(
		currentPrice decimal.Decimal,
		takeProfitPrice decimal.Decimal,
		stopLossPrice decimal.Decimal,
		expirationPeriod []Candlestick,
	) (OrderResult, time.Time)
	CalculateHeight(candlesticks []Candlestick) decimal.Decimal
	CalculateDepth(candlesticks []Candlestick) decimal.Decimal
	CalculateIsRising(candlesticks []Candlestick, countLast int, direction int) bool
}

type calculator struct {
}

func NewDefaultCalculator() Calculator {
	return &calculator{}
}

func (r calculator) CalculateMaxChange(candlesticks []Candlestick) decimal.Decimal {
	return r.CalculateHeight(candlesticks).Sub(r.CalculateDepth(candlesticks))
}

func (r calculator) CalculateVolume(candlesticks []Candlestick) decimal.Decimal {
	volume := decimal.NewFromInt(0)
	for i := range candlesticks {
		volume = volume.Add(decimal.NewFromInt(int64(candlesticks[i].Volume)))
	}

	return volume
}

func (r calculator) CalculateOrderResult(
	currentPrice decimal.Decimal,
	takeProfitPrice decimal.Decimal,
	stopLossPrice decimal.Decimal,
	expirationPeriod []Candlestick,
) (OrderResult, time.Time) {
	if len(expirationPeriod) == 0 {
		return OrderResultExpired, time.Now()
	}

	var direction = 0
	if currentPrice.LessThan(takeProfitPrice) {
		direction = 1
	}
	if currentPrice.GreaterThan(takeProfitPrice) {
		direction = -1
	}
	for _, c := range expirationPeriod {
		if direction >= 0 && c.Low.LessThanOrEqual(stopLossPrice) {
			return OrderResultStopLoss, c.Timestamp
		}
		if direction >= 0 && c.High.GreaterThanOrEqual(takeProfitPrice) {
			return OrderResultTakeProfit, c.Timestamp
		}
		if direction < 0 && c.High.GreaterThanOrEqual(stopLossPrice) {
			return OrderResultStopLoss, c.Timestamp
		}
		if direction < 0 && c.Low.LessThanOrEqual(takeProfitPrice) {
			return OrderResultTakeProfit, c.Timestamp
		}
	}

	return OrderResultExpired, expirationPeriod[len(expirationPeriod)-1].Timestamp
}

func (r calculator) CalculateVolatility(candlesticks []Candlestick) decimal.Decimal {
	if len(candlesticks) == 0 {
		return decimal.NewFromInt(0)
	}

	count := decimal.NewFromInt(int64(len(candlesticks)))
	avgClosePrice := r.CalculateSMA(candlesticks)
	squareDivsSum := decimal.NewFromInt(0)
	for i := range candlesticks {
		squareDivsSum = squareDivsSum.Add(avgClosePrice.Sub(candlesticks[i].High).Pow(decimal.NewFromInt(2)))
		squareDivsSum = squareDivsSum.Add(avgClosePrice.Sub(candlesticks[i].Low).Pow(decimal.NewFromInt(2)))
	}

	f, _ := squareDivsSum.Div(count.Mul(decimal.NewFromInt(2))).Float64()
	return decimal.NewFromFloat(math.Pow(f, 0.5))
}

func (r calculator) CalculateSMA(candlesticks []Candlestick) decimal.Decimal {
	if len(candlesticks) == 0 {
		return decimal.NewFromInt(0)
	}

	count := decimal.NewFromInt(int64(len(candlesticks)))
	closePricesSum := decimal.NewFromInt(0)
	for i := range candlesticks {
		closePricesSum = closePricesSum.Add(candlesticks[i].Close)
	}

	return closePricesSum.Div(count)
}

func (r calculator) CalculateHeight(candlesticks []Candlestick) decimal.Decimal {
	if len(candlesticks) == 0 {
		return decimal.NewFromInt(0)
	}

	height := candlesticks[0].High
	for i := range candlesticks {
		if height.LessThan(candlesticks[i].High) {
			height = candlesticks[i].High
		}
	}

	return height
}

func (r calculator) CalculateDepth(candlesticks []Candlestick) decimal.Decimal {
	if len(candlesticks) == 0 {
		return decimal.NewFromInt(0)
	}

	depth := candlesticks[0].Low
	for i := range candlesticks {
		if depth.GreaterThan(candlesticks[i].Low) {
			depth = candlesticks[i].Low
		}
	}

	return depth
}

func (r calculator) CalculateIsRising(candlesticks []Candlestick, countLast int, direction int) bool {
	for i := len(candlesticks) - 1; i >= 0; i-- {
		if (direction > 0 && candlesticks[i].Close.LessThan(candlesticks[i].Open)) ||
			(direction < 0 && candlesticks[i].Close.GreaterThan(candlesticks[i].Open)) {
			return false
		}
		countLast--
		if countLast <= 0 {
			break
		}
	}

	return true
}
