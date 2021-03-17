package candlestick

import (
	"math"

	"github.com/shopspring/decimal"
)

type OrderResult int

const (
	_ OrderResult = iota
	OrderResultProfitBuy
	OrderResultProfitSell
	OrderResultStopLoss
	OrderResultExpired
)

type Calculator interface {
	CalculateSMA(candlesticks []Candlestick) decimal.Decimal
	CalculateVolatility(candlesticks []Candlestick) decimal.Decimal
	CalculateRelativeVolatility(candlesticks []Candlestick) decimal.Decimal
	CalculateVolume(candlesticks []Candlestick) decimal.Decimal
	CalculateOrderResult(
		currentPrice decimal.Decimal,
		takeProfitPriceDiff decimal.Decimal,
		stopLossPriceDiff decimal.Decimal,
		expirationPeriod []Candlestick,
	) (result OrderResult, closeHour decimal.Decimal)
}

type calculator struct {
}

func DefaultCalculator() Calculator {
	return &calculator{}
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
	takeProfitPriceDiff decimal.Decimal,
	stopLossPriceDiff decimal.Decimal,
	expirationPeriod []Candlestick,
) (result OrderResult, closeHour decimal.Decimal) {
	var buyStopLoss, sellStopLoss bool
	closeHour = decimal.NewFromInt(0)
	for _, c := range expirationPeriod {
		closeHour = closeHour.Add(decimal.NewFromInt(1))
		if c.High.GreaterThanOrEqual(currentPrice.Add(takeProfitPriceDiff)) && !buyStopLoss {
			return OrderResultProfitBuy, closeHour
		}
		if c.Low.LessThanOrEqual(currentPrice.Sub(stopLossPriceDiff)) {
			buyStopLoss = true
		}
		if c.Low.LessThanOrEqual(currentPrice.Sub(takeProfitPriceDiff)) && !sellStopLoss {
			return OrderResultProfitSell, closeHour
		}
		if c.High.GreaterThanOrEqual(currentPrice.Add(stopLossPriceDiff)) {
			sellStopLoss = true
		}
	}

	if buyStopLoss || sellStopLoss {
		return OrderResultStopLoss, closeHour
	}

	return OrderResultExpired, closeHour
}

func (r calculator) CalculateVolatility(candlesticks []Candlestick) decimal.Decimal {
	if len(candlesticks) == 0 {
		return decimal.NewFromInt(0)
	}

	count := decimal.NewFromInt(int64(len(candlesticks)))
	avgClosePrice := r.CalculateSMA(candlesticks)
	squareDivsSum := decimal.NewFromInt(0)
	for i := range candlesticks {
		squareDivsSum = squareDivsSum.Add(avgClosePrice.Sub(candlesticks[i].Open).Pow(decimal.NewFromInt(2)))
		squareDivsSum = squareDivsSum.Add(avgClosePrice.Sub(candlesticks[i].Close).Pow(decimal.NewFromInt(2)))
		squareDivsSum = squareDivsSum.Add(avgClosePrice.Sub(candlesticks[i].High).Pow(decimal.NewFromInt(2)))
		squareDivsSum = squareDivsSum.Add(avgClosePrice.Sub(candlesticks[i].Low).Pow(decimal.NewFromInt(2)))
	}

	f, _ := squareDivsSum.Div(count.Mul(decimal.NewFromInt(4))).Float64()
	return decimal.NewFromFloat(math.Pow(f, 0.5))
}

func (r calculator) CalculateRelativeVolatility(candlesticks []Candlestick) decimal.Decimal {
	if len(candlesticks) == 0 {
		return decimal.NewFromInt(0)
	}

	return r.CalculateVolatility(candlesticks).Div(r.CalculateSMA(candlesticks)).Mul(decimal.NewFromInt(100))
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
