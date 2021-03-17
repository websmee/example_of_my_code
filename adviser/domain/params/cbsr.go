package params

import "github.com/shopspring/decimal"

type CBSR struct {
	PeriodHoursMin                 int
	PeriodHoursMax                 int
	StormToCalmMin                 decimal.Decimal
	StormToCalmMax                 decimal.Decimal
	StormMinPowerToCalmVolatility  decimal.Decimal
	StormMaxPowerToCalmVolatility  decimal.Decimal
	StormMinVolumeToCalmVolume     decimal.Decimal
	CalmMaxVolatilityToStormPower  decimal.Decimal
	CalmMaxCurvatureToStormPower   decimal.Decimal
	TakeProfitDiffToStormPower     decimal.Decimal
	StopLossDiffToStormPower       decimal.Decimal
	CalmToCheckDirection           decimal.Decimal
	StormPowerToCheckDirectionDiff decimal.Decimal
}

func (r *CBSR) GetParams() []decimal.Decimal {
	return []decimal.Decimal{
		decimal.NewFromInt(int64(r.PeriodHoursMin)),
		decimal.NewFromInt(int64(r.PeriodHoursMax)),
		r.StormToCalmMin,
		r.StormToCalmMax,
		r.StormMinPowerToCalmVolatility,
		r.StormMaxPowerToCalmVolatility,
		r.StormMinVolumeToCalmVolume,
		r.CalmMaxVolatilityToStormPower,
		r.CalmMaxCurvatureToStormPower,
		r.TakeProfitDiffToStormPower,
		r.StopLossDiffToStormPower,
		r.CalmToCheckDirection,
		r.StormPowerToCheckDirectionDiff,
	}
}

func (r *CBSR) SetParams(params []decimal.Decimal) {
	r.PeriodHoursMin = int(params[0].IntPart())
	r.PeriodHoursMax = int(params[1].IntPart())
	r.StormToCalmMin = params[2]
	r.StormToCalmMax = params[3]
	r.StormMinPowerToCalmVolatility = params[4]
	r.StormMaxPowerToCalmVolatility = params[5]
	r.StormMinVolumeToCalmVolume = params[6]
	r.CalmMaxVolatilityToStormPower = params[7]
	r.CalmMaxCurvatureToStormPower = params[8]
	r.TakeProfitDiffToStormPower = params[9]
	r.StopLossDiffToStormPower = params[10]
	r.CalmToCheckDirection = params[11]
	r.StormPowerToCheckDirectionDiff = params[12]
}
