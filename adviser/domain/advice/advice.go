package advice

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
	"github.com/websmee/example_of_my_code/adviser/domain/quote"
)

const DefaultLeverage = 1

type Status string

const (
	StatusOK Status = "ok"
)

type Advice struct {
	Quote            quote.Quote
	Candlesticks     []candlestick.Candlestick
	Price            decimal.Decimal
	Amount           decimal.Decimal
	TakeProfitPrice  decimal.Decimal
	TakeProfitAmount decimal.Decimal
	StopLossPrice    decimal.Decimal
	StopLossAmount   decimal.Decimal
	Leverage         int
	ExpiresAt        time.Time
}

type InternalAdvice struct {
	Status        Status
	QuoteSymbol   string
	HoursBefore   int
	HoursAfter    int
	Timestamp     time.Time
	CurrentPrice  decimal.Decimal
	TakeProfit    decimal.Decimal
	StopLoss      decimal.Decimal
	Leverage      int
	OrderResult   candlestick.OrderResult
	OrderClosed   time.Time
	AdviserType   AdviserType
	AdviserParams []decimal.Decimal
}
