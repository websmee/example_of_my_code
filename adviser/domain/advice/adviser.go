package advice

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
)

type AdviserType string

const (
	AdviserTypeCBS AdviserType = "CBS"
	AdviserTypeFT  AdviserType = "FT"
)

type Adviser interface {
	GetAdvices(ctx context.Context, adviserParams []decimal.Decimal, current candlestick.Candlestick, quoteSymbol string) ([]InternalAdvice, error)
}
