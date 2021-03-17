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
	FTOk
	FTTrendTooVolatile
	FTTrendTooStrong
	FTTrendTooWeak
	FTTrendWrongDirection
	FTSystemError
)

const ftExpirationHours = 72

type ftAdviser struct {
	repository candlestick.Repository
	calc       candlestick.Calculator
}

func NewFTAdviser(repository candlestick.Repository, calc candlestick.Calculator) Adviser {
	return &ftAdviser{
		repository: repository,
		calc:       calc,
	}
}

func (r ftAdviser) GetAdvice(
	ctx context.Context,
	adviserParams []decimal.Decimal,
	current candlestick.Candlestick,
	quoteSymbol string,
) (*Advice, NoAdviceReason, error) {
	ftParams := new(params.FT)
	ftParams.SetParams(adviserParams)

	var (
		reason         NoAdviceReason
		err            error
		trendDirection int64
	)

	reason, trendDirection, err = r.checkTrend(ctx, ftParams, current, quoteSymbol)
	if reason != FTOk {
		return nil, reason, err
	}

	expiration := current.Timestamp.Add(ftExpirationHours * time.Hour)

	var tp, sl decimal.Decimal
	if trendDirection > 0 {
		tp = current.Close.Add(ftParams.TakeProfitDiff)
		sl = current.Close.Sub(ftParams.StopLossDiff)
	} else {
		tp = current.Close.Sub(ftParams.TakeProfitDiff)
		sl = current.Close.Add(ftParams.StopLossDiff)
	}

	if reason, err := r.checkDirection(ctx, ftParams, current, quoteSymbol, trendDirection); reason != FTOk {
		return nil, reason, err
	}

	return &Advice{
		TakeProfit: tp,
		StopLoss:   sl,
		Expiration: expiration,
	}, FTOk, nil
}

func (r ftAdviser) GetReasonName(reason NoAdviceReason) string {
	switch reason {
	case FTOk:
		return "FTOk"
	case FTTrendTooVolatile:
		return "FTTrendTooVolatile"
	case FTTrendTooStrong:
		return "FTTrendTooStrong"
	case FTTrendTooWeak:
		return "FTTrendTooWeak"
	case FTTrendWrongDirection:
		return "FTTrendWrongDirection"
	case FTSystemError:
		return "FTSystemError"
	default:
		return "FTUnknown"
	}
}

func (r ftAdviser) checkTrend(
	ctx context.Context,
	ftParams *params.FT,
	current candlestick.Candlestick,
	quoteSymbol string,
) (NoAdviceReason, int64, error) {
	trend, err := r.repository.GetCandlesticks(
		ctx,
		quoteSymbol,
		candlestick.IntervalHour,
		current.Timestamp.Add(-time.Duration(ftParams.TrendDurationHours)*time.Hour),
		current.Timestamp,
	)
	if err != nil {
		return FTSystemError, 0, err
	}

	if r.calc.CalculateVolatility(trend).GreaterThan(ftParams.TrendMaxVolatility) {
		return FTTrendTooVolatile, 0, nil
	}

	if current.Close.Sub(trend[0].Open).Abs().GreaterThan(ftParams.TrendMaxCurvature) {
		return FTTrendTooStrong, 0, nil
	}

	if current.Close.Sub(trend[0].Open).Abs().LessThan(ftParams.TrendMinCurvature) {
		return FTTrendTooWeak, 0, nil
	}

	var direction int64 = 1 // up
	if trend[0].Open.GreaterThan(current.Close) {
		direction = -1 // down
	}

	return FTOk, direction, nil
}

func (r ftAdviser) checkDirection(
	ctx context.Context,
	ftParams *params.FT,
	current candlestick.Candlestick,
	quoteSymbol string,
	direction int64,
) (NoAdviceReason, error) {
	directionPeriod, err := r.repository.GetCandlesticks(
		ctx,
		quoteSymbol,
		candlestick.IntervalHour,
		current.Timestamp.Add(-time.Duration(ftParams.CheckDirectionHours)*24*time.Hour),
		current.Timestamp,
	)
	if err != nil {
		return FTSystemError, err
	}

	sma := r.calc.CalculateSMA(directionPeriod)
	if (sma.Sub(current.Close).GreaterThan(ftParams.CheckDirectionDiff) && direction < 0) ||
		(sma.Sub(current.Close).LessThan(ftParams.CheckDirectionDiff.Neg()) && direction > 0) {
		return FTTrendWrongDirection, nil
	}

	return FTOk, nil
}
