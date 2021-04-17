package advice

import "github.com/shopspring/decimal"

type CBSParams struct {
	CalmDurationHours   int
	CalmMaxChange       decimal.Decimal
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

func (r *CBSParams) GetParams() []decimal.Decimal {
	return []decimal.Decimal{
		decimal.NewFromInt(int64(r.CalmDurationHours)),
		r.CalmMaxChange,
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

func (r *CBSParams) SetParams(params []decimal.Decimal) {
	r.CalmDurationHours = int(params[0].IntPart())
	r.CalmMaxChange = params[1]
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
