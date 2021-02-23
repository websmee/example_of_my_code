package params

import "github.com/shopspring/decimal"

type Repository interface {
	SaveParams(name string, params []decimal.Decimal) error
	LoadParams(name string) ([]decimal.Decimal, error)
}
