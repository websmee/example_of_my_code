package advice

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
	"github.com/websmee/example_of_my_code/adviser/domain/signal"
)

type NoAdviceReason int

const (
	_ NoAdviceReason = iota
	CBSOk
	CBSTooShortCalmPeriod
	CBSNotCalmEnough
	CBSTooSteepCalmLine
	CBSTooWeakStorm
	CBSTooStrongStorm
	CBSSystemError
)

type Adviser interface {
	GetAdvice(ctx context.Context, signalParams []decimal.Decimal, current candlestick.Candlestick, quoteSymbol string) (*Advice, NoAdviceReason, error)
}

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
	signalParams []decimal.Decimal,
	current candlestick.Candlestick,
	quoteSymbol string,
) (*Advice, NoAdviceReason, error) {
	cbsParams := new(signal.CBSParams)
	cbsParams.SetParams(signalParams)

	if reason, err := r.checkCalm(ctx, cbsParams, current, quoteSymbol); reason != CBSOk {
		return nil, reason, err
	}

	reason, direction, err := r.checkStorm(ctx, cbsParams, current, quoteSymbol)
	if reason != CBSOk {
		return nil, reason, err
	}

	expiration := current.Timestamp.Add(time.Duration(cbsParams.StormDurationHours) * time.Hour)

	var tp, sl decimal.Decimal
	if direction > 0 {
		tp = current.Close.Add(cbsParams.TakeProfitDiff)
		sl = current.Close.Sub(cbsParams.StopLossDiff)
	} else {
		tp = current.Close.Sub(cbsParams.TakeProfitDiff)
		sl = current.Close.Add(cbsParams.StopLossDiff)
	}

	return &Advice{
		TakeProfit: tp,
		StopLoss:   sl,
		Expiration: expiration,
	}, CBSOk, nil
}

func (r cbsAdviser) checkCalm(
	ctx context.Context,
	cbsParams *signal.CBSParams,
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

	if len(calm) < cbsParams.CalmDurationHours/2 {
		return CBSTooShortCalmPeriod, nil
	}

	if r.calc.CalculateVolatility(calm).GreaterThan(cbsParams.CalmMaxVolatility) {
		return CBSNotCalmEnough, nil
	}

	if current.Close.Sub(r.calc.CalculateSMA(calm)).GreaterThan(cbsParams.CalmMaxCurvature) {
		return CBSTooSteepCalmLine, nil
	}

	return CBSOk, nil
}

func (r cbsAdviser) checkStorm(
	ctx context.Context,
	cbsParams *signal.CBSParams,
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
		return CBSTooWeakStorm, 0, nil
	}

	if stormPower.GreaterThan(cbsParams.StormMaxPower) {
		return CBSTooStrongStorm, 0, nil
	}

	var direction int64 = 1 // up
	if storm[0].Open.GreaterThan(current.Close) {
		direction = -1 // down
	}

	return CBSOk, direction, nil
}
