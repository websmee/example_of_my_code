package grpc

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/websmee/example_of_my_code/quotes/api/proto"

	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
)

func (r quotesAppGRPCClient) GetCandlesticks(ctx context.Context, symbol string, interval candlestick.Interval, from, to time.Time) ([]candlestick.Candlestick, error) {
	resp, err := r.getCandlesticksEndpoint(ctx, GetCandlesticksRequest{
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

func decodeGRPCGetCandlesticksResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*proto.GetCandlesticksReply)
	cs := make([]candlestick.Candlestick, len(reply.Candlesticks))
	for i := range reply.Candlesticks {
		cs[i] = candlestick.Candlestick{
			Open:      decimal.NewFromFloat32(reply.Candlesticks[i].Open),
			Low:       decimal.NewFromFloat32(reply.Candlesticks[i].Low),
			High:      decimal.NewFromFloat32(reply.Candlesticks[i].High),
			Close:     decimal.NewFromFloat32(reply.Candlesticks[i].Close),
			AdjClose:  decimal.NewFromFloat32(reply.Candlesticks[i].AdjClose),
			Volume:    int(reply.Candlesticks[i].Volume),
			Timestamp: time.Unix(reply.Candlesticks[i].Timestamp, 0),
			Interval:  candlestick.Interval(reply.Candlesticks[i].Interval),
			QuoteID:   reply.Candlesticks[i].QuoteId,
		}
	}

	return GetCandlesticksResponse{
		Candlesticks: cs,
		Err:          nil,
	}, nil
}

func encodeGRPCGetCandlesticksRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(GetCandlesticksRequest)
	return &proto.GetCandlesticksRequest{
		Symbol:   req.Symbol,
		Interval: string(req.Interval),
		From:     req.From.Format(time.RFC3339),
		To:       req.To.Format(time.RFC3339),
	}, nil
}
