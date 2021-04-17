package app

import (
	"github.com/go-kit/kit/log"
)

type CandlestickLoaderMiddleware func(service CandlestickLoader) CandlestickLoader

func CandlestickLoaderLoggingMiddleware(logger log.Logger) CandlestickLoaderMiddleware {
	return func(next CandlestickLoader) CandlestickLoader {
		return candlestickLoaderLoggingMiddleware{logger, next}
	}
}

type candlestickLoaderLoggingMiddleware struct {
	logger log.Logger
	next   CandlestickLoader
}

func (mw candlestickLoaderLoggingMiddleware) LoadCandlesticks() (err error) {
	defer func() {
		_ = mw.logger.Log("method", "LoadCandlesticks", "error", err)
	}()
	err = mw.next.LoadCandlesticks()
	return
}

func (mw candlestickLoaderLoggingMiddleware) LoadLatest() (err error) {
	defer func() {
		_ = mw.logger.Log("method", "LoadLatest", "error", err)
	}()
	err = mw.next.LoadLatest()
	return
}
