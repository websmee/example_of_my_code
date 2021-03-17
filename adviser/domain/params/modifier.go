package params

import (
	"github.com/shopspring/decimal"
)

type Modifier interface {
	Modify(modifying []decimal.Decimal) (stop bool)
	GetCurrentStep() int
	GetTotalSteps() int
}
