package params

import "github.com/shopspring/decimal"

type CBS struct {
	CalmDurationHours   int
	CalmMaxVolatility   decimal.Decimal
	CalmMaxCurvature    decimal.Decimal
	StormDurationHours  int
	StormMinPower       decimal.Decimal
	StormMaxPower       decimal.Decimal
	StormMinVolume      decimal.Decimal
	TakeProfitDiff      decimal.Decimal
	StopLossDiff        decimal.Decimal
	CheckDirectionHours int
	CheckDirectionDiff  decimal.Decimal
}

func (r *CBS) GetParams() []decimal.Decimal {
	return []decimal.Decimal{
		decimal.NewFromInt(int64(r.CalmDurationHours)),
		r.CalmMaxVolatility,
		r.CalmMaxCurvature,
		decimal.NewFromInt(int64(r.StormDurationHours)),
		r.StormMinPower,
		r.StormMaxPower,
		r.StormMinVolume,
		r.TakeProfitDiff,
		r.StopLossDiff,
		decimal.NewFromInt(int64(r.CheckDirectionHours)),
		r.CheckDirectionDiff,
	}
}

func (r *CBS) SetParams(params []decimal.Decimal) {
	r.CalmDurationHours = int(params[0].IntPart())
	r.CalmMaxVolatility = params[1]
	r.CalmMaxCurvature = params[2]
	r.StormDurationHours = int(params[3].IntPart())
	r.StormMinPower = params[4]
	r.StormMaxPower = params[5]
	r.StormMinVolume = params[6]
	r.TakeProfitDiff = params[7]
	r.StopLossDiff = params[8]
	r.CheckDirectionHours = int(params[9].IntPart())
	r.CheckDirectionDiff = params[10]
}
