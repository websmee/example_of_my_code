package grpc

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/pkg/errors"

	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
)

type CandlestickRepository struct {
	GetCandlesticksEndpoint endpoint.Endpoint
}

func (r CandlestickRepository) GetCandlesticks(ctx context.Context, symbol string, interval candlestick.Interval, from, to time.Time) ([]candlestick.Candlestick, error) {
	resp, err := r.GetCandlesticksEndpoint(ctx, GetCandlesticksRequest{
		Symbol:   symbol,
		Interval: interval,
		From:     from,
		To:       to,
	})
	if err != nil {
		return nil, errors.Wrap(err, "GetCandlesticks failed")
	}

	return resp.(GetCandlesticksResponse).Candlesticks, nil
}

var (
	_ endpoint.Failer = GetCandlesticksResponse{}
)

type GetCandlesticksRequest struct {
	Symbol   string
	Interval candlestick.Interval
	From, To time.Time
}

type GetCandlesticksResponse struct {
	Candlesticks []candlestick.Candlestick
	Err          error
}

func (r GetCandlesticksResponse) Failed() error { return r.Err }
