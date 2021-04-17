package advice

import (
	"github.com/shopspring/decimal"
)

type FTParams struct {
	TrendDurationHours  int
	TrendMaxVolatility  decimal.Decimal
	TrendMinCurvature   decimal.Decimal
	TrendMaxCurvature   decimal.Decimal
	TakeProfitDiff      decimal.Decimal
	StopLossDiff        decimal.Decimal
	CheckDirectionHours int
	CheckDirectionDiff  decimal.Decimal
}

func (r *FTParams) GetParams() []decimal.Decimal {
	return []decimal.Decimal{
		decimal.NewFromInt(int64(r.TrendDurationHours)),
		r.TrendMaxVolatility,
		r.TrendMinCurvature,
		r.TrendMaxCurvature,
		r.TakeProfitDiff,
		r.StopLossDiff,
		decimal.NewFromInt(int64(r.CheckDirectionHours)),
		r.CheckDirectionDiff,
	}
}

func (r *FTParams) SetParams(params []decimal.Decimal) {
	r.TrendDurationHours = int(params[0].IntPart())
	r.TrendMaxVolatility = params[1]
	r.TrendMinCurvature = params[2]
	r.TrendMaxCurvature = params[3]
	r.TakeProfitDiff = params[4]
	r.StopLossDiff = params[5]
	r.CheckDirectionHours = int(params[6].IntPart())
	r.CheckDirectionDiff = params[7]
}
