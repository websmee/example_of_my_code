package signal

import (
	"github.com/shopspring/decimal"
)

type CBSParams struct {
	CalmDurationHours  int
	CalmMaxVolatility  decimal.Decimal
	CalmMaxCurvature   decimal.Decimal
	StormDurationHours int
	StormMinPower      decimal.Decimal
	StormMaxPower      decimal.Decimal
	TakeProfitDiff     decimal.Decimal
	StopLossDiff       decimal.Decimal
}

func (r *CBSParams) GetParams() []decimal.Decimal {
	return []decimal.Decimal{
		decimal.NewFromInt(int64(r.CalmDurationHours)),
		r.CalmMaxVolatility,
		r.CalmMaxCurvature,
		decimal.NewFromInt(int64(r.StormDurationHours)),
		r.StormMinPower,
		r.StormMaxPower,
		r.TakeProfitDiff,
		r.StopLossDiff,
	}
}

func (r *CBSParams) SetParams(params []decimal.Decimal) {
	r.CalmDurationHours = int(params[0].IntPart())
	r.CalmMaxVolatility = params[1]
	r.CalmMaxCurvature = params[2]
	r.StormDurationHours = int(params[3].IntPart())
	r.StormMinPower = params[4]
	r.StormMaxPower = params[5]
	r.TakeProfitDiff = params[6]
	r.StopLossDiff = params[7]
}
