package advice

import "github.com/shopspring/decimal"

type CBSScaledParams struct {
	PeriodHoursMin                 int
	PeriodHoursMax                 int
	StormToCalmMin                 decimal.Decimal
	StormToCalmMax                 decimal.Decimal
	StormMinPowerToCalmMaxChange   decimal.Decimal
	StormMaxPowerToCalmMaxChange   decimal.Decimal
	StormMinVolumeToCalmVolume     decimal.Decimal
	CalmMaxChangeToStormPower      decimal.Decimal
	CalmMaxCurvatureToStormPower   decimal.Decimal
	TakeProfitDiffToStormPower     decimal.Decimal
	StopLossDiffToStormPower       decimal.Decimal
	CalmToCheckDirection           decimal.Decimal
	StormPowerToCheckDirectionDiff decimal.Decimal
}

func (r *CBSScaledParams) GetParams() []decimal.Decimal {
	return []decimal.Decimal{
		decimal.NewFromInt(int64(r.PeriodHoursMin)),
		decimal.NewFromInt(int64(r.PeriodHoursMax)),
		r.StormToCalmMin,
		r.StormToCalmMax,
		r.StormMinPowerToCalmMaxChange,
		r.StormMaxPowerToCalmMaxChange,
		r.StormMinVolumeToCalmVolume,
		r.CalmMaxChangeToStormPower,
		r.CalmMaxCurvatureToStormPower,
		r.TakeProfitDiffToStormPower,
		r.StopLossDiffToStormPower,
		r.CalmToCheckDirection,
		r.StormPowerToCheckDirectionDiff,
	}
}

func (r *CBSScaledParams) SetParams(params []decimal.Decimal) {
	r.PeriodHoursMin = int(params[0].IntPart())
	r.PeriodHoursMax = int(params[1].IntPart())
	r.StormToCalmMin = params[2]
	r.StormToCalmMax = params[3]
	r.StormMinPowerToCalmMaxChange = params[4]
	r.StormMaxPowerToCalmMaxChange = params[5]
	r.StormMinVolumeToCalmVolume = params[6]
	r.CalmMaxChangeToStormPower = params[7]
	r.CalmMaxCurvatureToStormPower = params[8]
	r.TakeProfitDiffToStormPower = params[9]
	r.StopLossDiffToStormPower = params[10]
	r.CalmToCheckDirection = params[11]
	r.StormPowerToCheckDirectionDiff = params[12]
}
