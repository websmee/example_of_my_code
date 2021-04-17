package advice

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
)

type cbsScaledAdviser struct {
	cbsAdviser *cbsAdviser
	repository candlestick.Repository
	calc       candlestick.Calculator
}

func NewCBSScaledAdviser(repository candlestick.Repository, calc candlestick.Calculator) Adviser {
	return &cbsScaledAdviser{
		cbsAdviser: NewCBSAdviser(repository, calc).(*cbsAdviser),
		repository: repository,
		calc:       calc,
	}
}

func (r cbsScaledAdviser) GetAdvices(
	ctx context.Context,
	adviserParams []decimal.Decimal,
	current candlestick.Candlestick,
	quoteSymbol string,
) ([]InternalAdvice, error) {
	cbsScaledParams := new(CBSScaledParams)
	cbsScaledParams.SetParams(adviserParams)

	var advices []InternalAdvice

	for p := cbsScaledParams.PeriodHoursMax; p >= cbsScaledParams.PeriodHoursMin; p-- {
		stormHoursMin := decimal.NewFromInt(int64(p)).Mul(cbsScaledParams.StormToCalmMin).IntPart()
		stormHoursMax := decimal.NewFromInt(int64(p)).Mul(cbsScaledParams.StormToCalmMax).IntPart()
		stormHours := stormHoursMax
		calmHours := int64(p) - stormHours
		for stormHours > stormHoursMin {
			calmMaxChange, calmVolume, err := r.getCalmProps(ctx, current, quoteSymbol, calmHours, stormHours)
			if err != nil {
				return nil, err
			}

			stormPower, err := r.getStormPower(ctx, current, quoteSymbol, stormHours)
			if err != nil {
				return nil, err
			}

			cbsParams := CBSParams{
				CalmDurationHours:   int(calmHours),
				CalmMaxChange:       stormPower.Mul(cbsScaledParams.CalmMaxChangeToStormPower),
				CalmMaxCurvature:    stormPower.Mul(cbsScaledParams.CalmMaxCurvatureToStormPower),
				StormDurationHours:  int(stormHours),
				StormMinPower:       cbsScaledParams.StormMinPowerToCalmMaxChange.Mul(calmMaxChange),
				StormMaxPower:       cbsScaledParams.StormMaxPowerToCalmMaxChange.Mul(calmMaxChange),
				StormMinVolume:      calmVolume.Mul(cbsScaledParams.StormMinVolumeToCalmVolume),
				TakeProfitDiff:      stormPower.Mul(cbsScaledParams.TakeProfitDiffToStormPower),
				StopLossDiff:        stormPower.Mul(cbsScaledParams.StopLossDiffToStormPower),
				CheckDirectionHours: int(decimal.NewFromInt(calmHours).Div(cbsScaledParams.CalmToCheckDirection).IntPart()),
				CheckDirectionDiff:  stormPower.Div(cbsScaledParams.StormPowerToCheckDirectionDiff),
			}

			a, err := r.cbsAdviser.GetAdvices(ctx, cbsParams.GetParams(), current, quoteSymbol)
			if err != nil {
				return nil, err
			}
			calmHours++
			stormHours--

			advices = append(advices, a...)
		}
	}

	return advices, nil
}

func (r cbsScaledAdviser) getCalmProps(ctx context.Context, current candlestick.Candlestick, quoteSymbol string, calmHours, stormHours int64) (decimal.Decimal, decimal.Decimal, error) {
	calm, err := r.repository.GetCandlesticksByCount(
		ctx,
		quoteSymbol,
		candlestick.IntervalHour,
		current.Timestamp.Add(-time.Duration(stormHours)*time.Hour),
		candlestick.GetterDirectionBackward,
		int(calmHours),
	)
	if err != nil {
		return decimal.NewFromInt(0), decimal.NewFromInt(0), err
	}

	return r.calc.CalculateMaxChange(calm), r.calc.CalculateVolume(calm), nil
}

func (r cbsScaledAdviser) getStormPower(ctx context.Context, current candlestick.Candlestick, quoteSymbol string, hours int64) (decimal.Decimal, error) {
	storm, err := r.repository.GetCandlesticksByCount(
		ctx,
		quoteSymbol,
		candlestick.IntervalHour,
		current.Timestamp,
		candlestick.GetterDirectionBackward,
		int(hours),
	)
	if err != nil {
		return decimal.NewFromInt(0), err
	}

	return storm[0].Open.Sub(storm[len(storm)-1].Close).Abs(), nil
}
