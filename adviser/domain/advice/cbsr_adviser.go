package advice

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
	"github.com/websmee/example_of_my_code/adviser/domain/params"
)

const (
	_ NoAdviceReason = iota
	CBSROk
	CBSRCalmPeriodTooVolatile
	CBSRTooSteepCalmLine
	CBSRStormTooWeak
	CBSRStormTooStrong
	CBSRStormTooSmall
	CBSRWrongDirection
	CBSRSystemError
)

type cbsrAdviser struct {
	cbsAdviser *cbsAdviser
	repository candlestick.Repository
	calc       candlestick.Calculator
}

func NewCBSRAdviser(repository candlestick.Repository, calc candlestick.Calculator) Adviser {
	return &cbsrAdviser{
		cbsAdviser: NewCBSAdviser(repository, calc).(*cbsAdviser),
		repository: repository,
		calc:       calc,
	}
}

func (r cbsrAdviser) GetAdvice(
	ctx context.Context,
	adviserParams []decimal.Decimal,
	current candlestick.Candlestick,
	quoteSymbol string,
) (*Advice, NoAdviceReason, error) {
	cbsrParams := new(params.CBSR)
	cbsrParams.SetParams(adviserParams)

	reasons := make(map[NoAdviceReason]int)
	for p := cbsrParams.PeriodHoursMax; p >= cbsrParams.PeriodHoursMin; p-- {
		stormHoursMin := decimal.NewFromInt(int64(p)).Mul(cbsrParams.StormToCalmMin).IntPart()
		stormHoursMax := decimal.NewFromInt(int64(p)).Mul(cbsrParams.StormToCalmMax).IntPart()
		stormHours := stormHoursMax
		calmHours := int64(p) - stormHours
		for stormHours > stormHoursMin {
			calmVolatility, calmVolume, err := r.getCalmProps(ctx, current, quoteSymbol, calmHours, stormHours)
			if err != nil {
				return nil, CBSRSystemError, err
			}

			stormPower, err := r.getStormPower(ctx, current, quoteSymbol, stormHours)
			if err != nil {
				return nil, CBSRSystemError, err
			}

			cbsParams := params.CBS{
				CalmDurationHours:   int(calmHours),
				CalmMaxVolatility:   stormPower.Mul(cbsrParams.CalmMaxVolatilityToStormPower),
				CalmMaxCurvature:    stormPower.Mul(cbsrParams.CalmMaxCurvatureToStormPower),
				StormDurationHours:  int(stormHours),
				StormMinPower:       cbsrParams.StormMinPowerToCalmVolatility.Mul(calmVolatility),
				StormMaxPower:       cbsrParams.StormMaxPowerToCalmVolatility.Mul(calmVolatility),
				StormMinVolume:      calmVolume.Mul(cbsrParams.StormMinVolumeToCalmVolume),
				TakeProfitDiff:      stormPower.Mul(cbsrParams.TakeProfitDiffToStormPower),
				StopLossDiff:        stormPower.Mul(cbsrParams.StopLossDiffToStormPower),
				CheckDirectionHours: int(decimal.NewFromInt(calmHours).Div(cbsrParams.CalmToCheckDirection).IntPart()),
				CheckDirectionDiff:  stormPower.Div(cbsrParams.StormPowerToCheckDirectionDiff),
			}

			a, reason, err := r.cbsAdviser.GetAdvice(ctx, cbsParams.GetParams(), current, quoteSymbol)
			if reason != CBSOk {
				if reason == CBSSystemError {
					return nil, CBSRSystemError, err
				}
				reasons[reason]++
				calmHours++
				stormHours--
				continue
			}

			return a, CBSROk, nil
		}
	}

	var maxCountReason NoAdviceReason = -1
	maxCount := 0
	for rs, count := range reasons {
		if count > maxCount {
			maxCountReason = rs
		}
	}

	return nil, maxCountReason, nil
}

func (r cbsrAdviser) GetReasonName(reason NoAdviceReason) string {
	switch reason {
	case CBSROk:
		return "CBSROk"
	case CBSRCalmPeriodTooVolatile:
		return "CBSRCalmPeriodTooVolatile"
	case CBSRTooSteepCalmLine:
		return "CBSRTooSteepCalmLine"
	case CBSRStormTooWeak:
		return "CBSRStormTooWeak"
	case CBSRStormTooStrong:
		return "CBSRStormTooStrong"
	case CBSRStormTooSmall:
		return "CBSRStormTooSmall"
	case CBSRWrongDirection:
		return "CBSRWrongDirection"
	case CBSRSystemError:
		return "CBSRSystemError"
	default:
		return "CBSRUnknown"
	}
}

func (r cbsrAdviser) CBStoCBSRReason(reason NoAdviceReason) NoAdviceReason {
	switch reason {
	case CBSOk:
		return CBSROk
	case CBSCalmPeriodTooVolatile:
		return CBSRCalmPeriodTooVolatile
	case CBSTooSteepCalmLine:
		return CBSRTooSteepCalmLine
	case CBSStormTooWeak:
		return CBSRStormTooWeak
	case CBSStormTooStrong:
		return CBSRStormTooStrong
	case CBSStormTooSmall:
		return CBSRStormTooSmall
	case CBSWrongDirection:
		return CBSRWrongDirection
	case CBSSystemError:
		return CBSRSystemError
	default:
		return -1
	}
}

func (r cbsrAdviser) getCalmProps(ctx context.Context, current candlestick.Candlestick, quoteSymbol string, calmHours, stormHours int64) (decimal.Decimal, decimal.Decimal, error) {
	calm, err := r.repository.GetCandlesticks(
		ctx,
		quoteSymbol,
		candlestick.IntervalHour,
		current.Timestamp.Add(-time.Duration(calmHours+stormHours)*time.Hour),
		current.Timestamp.Add(-time.Duration(stormHours)*time.Hour),
	)
	if err != nil {
		return decimal.NewFromInt(0), decimal.NewFromInt(0), err
	}

	return r.calc.CalculateVolatility(calm), r.calc.CalculateVolume(calm), nil
}

func (r cbsrAdviser) getStormPower(ctx context.Context, current candlestick.Candlestick, quoteSymbol string, hours int64) (decimal.Decimal, error) {
	storm, err := r.repository.GetCandlesticks(
		ctx,
		quoteSymbol,
		candlestick.IntervalHour,
		current.Timestamp.Add(-time.Duration(hours)*time.Hour),
		current.Timestamp,
	)
	if err != nil {
		return decimal.NewFromInt(0), err
	}

	return storm[0].Open.Sub(storm[len(storm)-1].Close).Abs(), nil
}
