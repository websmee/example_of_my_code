package advice

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
)

const (
	StatusFTTrendTooVolatile    Status = "FTTrendTooVolatile"
	StatusFTTrendTooStrong      Status = "FTTrendTooStrong"
	StatusFTTrendTooWeak        Status = "FTTrendTooWeak"
	StatusFTTrendWrongDirection Status = "FTTrendWrongDirection"
)

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

func (r ftAdviser) GetAdvices(
	ctx context.Context,
	adviserParams []decimal.Decimal,
	current candlestick.Candlestick,
	quoteSymbol string,
) ([]InternalAdvice, error) {
	ftParams := new(FTParams)
	ftParams.SetParams(adviserParams)

	advices := []InternalAdvice{{
		Status:        StatusOK,
		QuoteSymbol:   quoteSymbol,
		HoursBefore:   ftParams.TrendDurationHours,
		HoursAfter:    ftParams.TrendDurationHours,
		Timestamp:     current.Timestamp,
		CurrentPrice:  current.Close,
		TakeProfit:    decimal.NewFromInt(0),
		StopLoss:      decimal.NewFromInt(0),
		Leverage:      DefaultLeverage,
		OrderResult:   candlestick.OrderResultNone,
		AdviserType:   AdviserTypeFT,
		AdviserParams: adviserParams,
	}}

	status, trendDirection, err := r.checkTrend(ctx, ftParams, current, quoteSymbol)
	if err != nil {
		return nil, err
	}
	if status != StatusOK {
		advices[0].Status = status
		return advices, err
	}

	var tp, sl decimal.Decimal
	if trendDirection > 0 {
		tp = current.Close.Add(ftParams.TakeProfitDiff)
		sl = current.Close.Sub(ftParams.StopLossDiff)
	} else {
		tp = current.Close.Sub(ftParams.TakeProfitDiff)
		sl = current.Close.Add(ftParams.StopLossDiff)
	}

	status, err = r.checkDirection(ctx, ftParams, current, quoteSymbol, trendDirection)
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

func (r ftAdviser) checkTrend(
	ctx context.Context,
	ftParams *FTParams,
	current candlestick.Candlestick,
	quoteSymbol string,
) (Status, int64, error) {
	trend, err := r.repository.GetCandlesticks(
		ctx,
		quoteSymbol,
		candlestick.IntervalHour,
		current.Timestamp.Add(-time.Duration(ftParams.TrendDurationHours)*time.Hour),
		current.Timestamp,
	)
	if err != nil {
		return "", 0, err
	}

	if r.calc.CalculateVolatility(trend).GreaterThan(ftParams.TrendMaxVolatility) {
		return StatusFTTrendTooVolatile, 0, nil
	}

	if current.Close.Sub(trend[0].Open).Abs().GreaterThan(ftParams.TrendMaxCurvature) {
		return StatusFTTrendTooStrong, 0, nil
	}

	if current.Close.Sub(trend[0].Open).Abs().LessThan(ftParams.TrendMinCurvature) {
		return StatusFTTrendTooWeak, 0, nil
	}

	var direction int64 = 1 // up
	if trend[0].Open.GreaterThan(current.Close) {
		direction = -1 // down
	}

	return StatusOK, direction, nil
}

func (r ftAdviser) checkDirection(
	ctx context.Context,
	ftParams *FTParams,
	current candlestick.Candlestick,
	quoteSymbol string,
	direction int64,
) (Status, error) {
	directionPeriod, err := r.repository.GetCandlesticks(
		ctx,
		quoteSymbol,
		candlestick.IntervalHour,
		current.Timestamp.Add(-time.Duration(ftParams.CheckDirectionHours)*24*time.Hour),
		current.Timestamp,
	)
	if err != nil {
		return "", err
	}

	sma := r.calc.CalculateSMA(directionPeriod)
	if (sma.Sub(current.Close).GreaterThan(ftParams.CheckDirectionDiff) && direction < 0) ||
		(sma.Sub(current.Close).LessThan(ftParams.CheckDirectionDiff.Neg()) && direction > 0) {
		return StatusFTTrendWrongDirection, nil
	}

	return StatusOK, nil
}
