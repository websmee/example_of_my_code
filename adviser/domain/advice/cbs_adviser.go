package advice

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
)

const (
	StatusCBSCalmPeriodIsUnstable Status = "CBSCalmPeriodIsUnstable"
	StatusCBSCalmPeriodTooShort   Status = "CBSCalmPeriodTooShort"
	StatusCBSTooSteepCalmLine     Status = "CBSTooSteepCalmLine"
	StatusCBSStormPeriodTooShort  Status = "CBSStormPeriodTooShort"
	StatusCBSStormTooWeak         Status = "CBSStormTooWeak"
	StatusCBSStormTooStrong       Status = "CBSStormTooStrong"
	StatusCBSStormTooSmall        Status = "CBSStormTooSmall"
	StatusCBSStormIsUnstable      Status = "CBSStormIsUnstable"
	StatusCBSWrongDirection       Status = "CBSWrongDirection"
)

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

func (r cbsAdviser) GetAdvices(
	ctx context.Context,
	adviserParams []decimal.Decimal,
	current candlestick.Candlestick,
	quoteSymbol string,
) ([]InternalAdvice, error) {
	cbsParams := new(CBSParams)
	cbsParams.SetParams(adviserParams)

	advices := []InternalAdvice{{
		Status:        StatusOK,
		QuoteSymbol:   quoteSymbol,
		HoursBefore:   cbsParams.CalmDurationHours + cbsParams.StormDurationHours,
		HoursAfter:    cbsParams.StormDurationHours,
		Timestamp:     current.Timestamp,
		CurrentPrice:  current.Close,
		TakeProfit:    decimal.NewFromInt(0),
		StopLoss:      decimal.NewFromInt(0),
		Leverage:      DefaultLeverage,
		OrderResult:   candlestick.OrderResultNone,
		AdviserType:   AdviserTypeCBS,
		AdviserParams: adviserParams,
	}}

	status, direction, err := r.checkStorm(ctx, cbsParams, current, quoteSymbol)
	if err != nil {
		return nil, err
	}
	if status != StatusOK {
		advices[0].Status = status
		return advices, err
	}

	status, err = r.checkCalm(ctx, cbsParams, current, quoteSymbol)
	if err != nil {
		return nil, err
	}
	if status != StatusOK {
		advices[0].Status = status
		return advices, err
	}

	var tp, sl decimal.Decimal
	if direction > 0 {
		tp = current.Close.Add(cbsParams.TakeProfitDiff)
		sl = current.Close.Sub(cbsParams.StopLossDiff)
	} else {
		tp = current.Close.Sub(cbsParams.TakeProfitDiff)
		sl = current.Close.Add(cbsParams.StopLossDiff)
	}

	status, err = r.checkDirection(ctx, cbsParams, current, quoteSymbol, direction)
	if err != nil {
		return nil, err
	}
	if status != StatusOK {
		advices[0].Status = status
		return advices, err
	}

	advices[0].TakeProfit = tp
	advices[0].StopLoss = sl

	return advices, nil
}

func (r cbsAdviser) checkCalm(
	ctx context.Context,
	cbsParams *CBSParams,
	current candlestick.Candlestick,
	quoteSymbol string,
) (Status, error) {
	calm, err := r.repository.GetCandlesticksByCount(
		ctx,
		quoteSymbol,
		candlestick.IntervalHour,
		current.Timestamp.Add(-time.Duration(cbsParams.StormDurationHours)*time.Hour),
		candlestick.GetterDirectionBackward,
		cbsParams.CalmDurationHours,
	)
	if err != nil {
		return "", err
	}

	if len(calm) < cbsParams.CalmDurationHours {
		return StatusCBSCalmPeriodTooShort, nil
	}

	if r.calc.CalculateMaxChange(calm).GreaterThan(cbsParams.CalmMaxChange) {
		return StatusCBSCalmPeriodIsUnstable, nil
	}

	if calm[0].Open.Sub(calm[len(calm)-1].Close).Abs().GreaterThan(cbsParams.CalmMaxCurvature) {
		return StatusCBSTooSteepCalmLine, nil
	}

	return StatusOK, nil
}

func (r cbsAdviser) checkStorm(
	ctx context.Context,
	cbsParams *CBSParams,
	current candlestick.Candlestick,
	quoteSymbol string,
) (Status, int, error) {
	storm, err := r.repository.GetCandlesticksByCount(
		ctx,
		quoteSymbol,
		candlestick.IntervalHour,
		current.Timestamp,
		candlestick.GetterDirectionBackward,
		cbsParams.StormDurationHours,
	)
	if err != nil {
		return "", 0, err
	}

	if len(storm) < cbsParams.StormDurationHours {
		return StatusCBSStormPeriodTooShort, 0, nil
	}

	stormPower := storm[0].Open.Sub(storm[len(storm)-1].Close).Abs()

	if stormPower.LessThan(cbsParams.StormMinPower) {
		return StatusCBSStormTooWeak, 0, nil
	}

	if stormPower.GreaterThan(cbsParams.StormMaxPower) {
		return StatusCBSStormTooStrong, 0, nil
	}

	stormVolume := r.calc.CalculateVolume(storm)
	if stormVolume.LessThan(cbsParams.StormMinVolume) {
		return StatusCBSStormTooSmall, 0, nil
	}

	direction := 1 // up
	if storm[0].Open.GreaterThan(current.Close) {
		direction = -1 // down
	}

	if !r.calc.CalculateIsRising(storm, cbsParams.StormDurationHours/5+1, direction) {
		return StatusCBSStormIsUnstable, 0, nil
	}

	return StatusOK, direction, nil
}

func (r cbsAdviser) checkDirection(
	ctx context.Context,
	cbsParams *CBSParams,
	current candlestick.Candlestick,
	quoteSymbol string,
	direction int,
) (Status, error) {
	directionPeriod, err := r.repository.GetCandlesticksByCount(
		ctx,
		quoteSymbol,
		candlestick.IntervalHour,
		current.Timestamp,
		candlestick.GetterDirectionBackward,
		cbsParams.CheckDirectionHours,
	)
	if err != nil {
		return "", err
	}

	sma := r.calc.CalculateSMA(directionPeriod)
	if (sma.Sub(current.Close).GreaterThan(cbsParams.CheckDirectionDiff) && direction < 0) ||
		(sma.Sub(current.Close).LessThan(cbsParams.CheckDirectionDiff.Neg()) && direction > 0) {
		return StatusCBSWrongDirection, nil
	}

	return StatusOK, nil
}
