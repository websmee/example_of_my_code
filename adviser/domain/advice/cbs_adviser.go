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
	CBSOk
	CBSCalmPeriodTooVolatile
	CBSTooSteepCalmLine
	CBSStormTooWeak
	CBSStormTooStrong
	CBSStormTooSmall
	CBSWrongDirection
	CBSSystemError
)

const cbsExpirationHoursCoeff = 3

type cbsAdviser struct {
	repository candlestick.Repository
	calc       candlestick.Calculator
}

func NewCBSAdviser(repository candlestick.Repository, calc candlestick.Calculator) Adviser {
	return &cbsAdviser{
		repository: repository,
		calc:       calc,
	}
}

func (r cbsAdviser) GetAdvice(
	ctx context.Context,
	adviserParams []decimal.Decimal,
	current candlestick.Candlestick,
	quoteSymbol string,
) (*Advice, NoAdviceReason, error) {
	cbsParams := new(params.CBS)
	cbsParams.SetParams(adviserParams)

	reason, direction, err := r.checkStorm(ctx, cbsParams, current, quoteSymbol)
	if reason != CBSOk {
		return nil, reason, err
	}

	if reason, err := r.checkCalm(ctx, cbsParams, current, quoteSymbol); reason != CBSOk {
		return nil, reason, err
	}

	expiration := current.Timestamp.Add(time.Duration(cbsParams.CalmDurationHours+cbsParams.StormDurationHours) * cbsExpirationHoursCoeff * time.Hour)

	var tp, sl decimal.Decimal
	if direction > 0 {
		tp = current.Close.Add(cbsParams.TakeProfitDiff)
		sl = current.Close.Sub(cbsParams.StopLossDiff)
	} else {
		tp = current.Close.Sub(cbsParams.TakeProfitDiff)
		sl = current.Close.Add(cbsParams.StopLossDiff)
	}

	if reason, err := r.checkDirection(ctx, cbsParams, current, quoteSymbol, direction); reason != CBSOk {
		return nil, reason, err
	}

	return &Advice{
		TakeProfit: tp,
		StopLoss:   sl,
		Expiration: expiration,
	}, CBSOk, nil
}

func (r cbsAdviser) GetReasonName(reason NoAdviceReason) string {
	switch reason {
	case CBSOk:
		return "CBSOk"
	case CBSCalmPeriodTooVolatile:
		return "CBSCalmPeriodTooVolatile"
	case CBSTooSteepCalmLine:
		return "CBSTooSteepCalmLine"
	case CBSStormTooWeak:
		return "CBSStormTooWeak"
	case CBSStormTooStrong:
		return "CBSStormTooStrong"
	case CBSStormTooSmall:
		return "CBSStormTooSmall"
	case CBSWrongDirection:
		return "CBSWrongDirection"
	case CBSSystemError:
		return "CBSSystemError"
	default:
		return "CBSUnknown"
	}
}

func (r cbsAdviser) checkCalm(
	ctx context.Context,
	cbsParams *params.CBS,
	current candlestick.Candlestick,
	quoteSymbol string,
) (NoAdviceReason, error) {
	calm, err := r.repository.GetCandlesticks(
		ctx,
		quoteSymbol,
		candlestick.IntervalHour,
		current.Timestamp.Add(-time.Duration(cbsParams.CalmDurationHours+cbsParams.StormDurationHours)*time.Hour),
		current.Timestamp.Add(-time.Duration(cbsParams.StormDurationHours)*time.Hour),
	)
	if err != nil {
		return CBSSystemError, err
	}

	if r.calc.CalculateVolatility(calm).GreaterThan(cbsParams.CalmMaxVolatility) {
		return CBSCalmPeriodTooVolatile, nil
	}

	if len(calm) == 0 {
		return CBSTooSteepCalmLine, nil
	}

	if calm[0].Open.Sub(calm[len(calm)-1].Close).Abs().GreaterThan(cbsParams.CalmMaxCurvature) {
		return CBSTooSteepCalmLine, nil
	}

	return CBSOk, nil
}

func (r cbsAdviser) checkStorm(
	ctx context.Context,
	cbsParams *params.CBS,
	current candlestick.Candlestick,
	quoteSymbol string,
) (NoAdviceReason, int64, error) {
	storm, err := r.repository.GetCandlesticks(
		ctx,
		quoteSymbol,
		candlestick.IntervalHour,
		current.Timestamp.Add(-time.Duration(cbsParams.StormDurationHours)*time.Hour),
		current.Timestamp,
	)
	if err != nil {
		return CBSSystemError, 0, err
	}

	stormPower := storm[0].Open.Sub(storm[len(storm)-1].Close).Abs()

	if stormPower.LessThan(cbsParams.StormMinPower) {
		return CBSStormTooWeak, 0, nil
	}

	if stormPower.GreaterThan(cbsParams.StormMaxPower) {
		return CBSStormTooStrong, 0, nil
	}

	stormVolume := r.calc.CalculateVolume(storm)
	if stormVolume.LessThan(cbsParams.StormMinVolume) {
		return CBSStormTooSmall, 0, nil
	}

	var direction int64 = 1 // up
	if storm[0].Open.GreaterThan(current.Close) {
		direction = -1 // down
	}

	return CBSOk, direction, nil
}

func (r cbsAdviser) checkDirection(
	ctx context.Context,
	cbsParams *params.CBS,
	current candlestick.Candlestick,
	quoteSymbol string,
	direction int64,
) (NoAdviceReason, error) {
	directionPeriod, err := r.repository.GetCandlesticks(
		ctx,
		quoteSymbol,
		candlestick.IntervalHour,
		current.Timestamp.Add(-time.Duration(cbsParams.CheckDirectionHours)*24*time.Hour),
		current.Timestamp,
	)
	if err != nil {
		return CBSSystemError, err
	}

	sma := r.calc.CalculateSMA(directionPeriod)
	if (sma.Sub(current.Close).GreaterThan(cbsParams.CheckDirectionDiff) && direction < 0) ||
		(sma.Sub(current.Close).LessThan(cbsParams.CheckDirectionDiff.Neg()) && direction > 0) {
		return CBSWrongDirection, nil
	}

	return CBSOk, nil
}
