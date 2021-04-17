package app

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"

	"github.com/websmee/example_of_my_code/adviser/domain/advice"
)

type AdviserMiddleware func(service AdviserApp) AdviserApp

func AdviserLoggingMiddleware(logger log.Logger) AdviserMiddleware {
	return func(next AdviserApp) AdviserApp {
		return adviserLoggingMiddleware{logger, next}
	}
}

type adviserLoggingMiddleware struct {
	logger log.Logger
	next   AdviserApp
}

func (mw adviserLoggingMiddleware) GetAdvices(ctx context.Context) (advices []advice.Advice, err error) {
	defer func() {
		_ = mw.logger.Log("method", "GetAdvices", "error", err)
	}()
	return mw.next.GetAdvices(ctx)
}

func (mw adviserLoggingMiddleware) HealthCheck() bool {
	return mw.next.HealthCheck()
}

func AdviserInstrumentingMiddleware(counter metrics.Counter) AdviserMiddleware {
	return func(next AdviserApp) AdviserApp {
		return adviserInstrumentingMiddleware{
			counter: counter,
			next:    next,
		}
	}
}

type adviserInstrumentingMiddleware struct {
	counter metrics.Counter
	next    AdviserApp
}

func (mw adviserInstrumentingMiddleware) GetAdvices(ctx context.Context) (advices []advice.Advice, err error) {
	advices, err = mw.next.GetAdvices(ctx)
	mw.counter.Add(float64(len(advices)))
	return
}

func (mw adviserInstrumentingMiddleware) HealthCheck() bool {
	return mw.next.HealthCheck()
}
