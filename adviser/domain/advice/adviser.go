package advice

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
)

type NoAdviceReason int

type Adviser interface {
	GetAdvice(ctx context.Context, adviserParams []decimal.Decimal, current candlestick.Candlestick, quoteSymbol string) (*Advice, NoAdviceReason, error)
	GetReasonName(reason NoAdviceReason) string
}
