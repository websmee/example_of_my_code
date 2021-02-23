package advice

import (
	"time"

	"github.com/shopspring/decimal"
)

type Advice struct {
	TakeProfit decimal.Decimal
	StopLoss   decimal.Decimal
	Expiration time.Time
}
