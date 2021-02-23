package api

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/quotes/api/proto"
	"github.com/websmee/example_of_my_code/quotes/domain/candlestick"
)

type grpcServer struct {
	getCandlesticks grpctransport.Handler
	proto.UnimplementedQuotesServer
}

func NewGRPCServer(endpoints Quotes, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) proto.QuotesServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	if zipkinTracer != nil {
		options = append(options, zipkin.GRPCServerTrace(zipkinTracer))
	}

	return &grpcServer{
		getCandlesticks: grpctransport.NewServer(
			endpoints.GetCandlesticksEndpoint,
			decodeGRPCGetCandlesticksRequest,
			encodeGRPCGetCandlesticksResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(otTracer, "GetCandlesticks", logger)))...,
		),
	}
}

func (s *grpcServer) GetCandlesticks(ctx context.Context, req *proto.GetCandlesticksRequest) (*proto.GetCandlesticksReply, error) {
	_, rep, err := s.getCandlesticks.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.GetCandlesticksReply), nil
}

func decodeGRPCGetCandlesticksRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	var from, to time.Time
	var err error

	req := grpcReq.(*proto.GetCandlesticksRequest)

	from, err = time.Parse(time.RFC3339, req.From)
	if err != nil {
		return nil, err
	}

	to, err = time.Parse(time.RFC3339, req.To)
	if err != nil {
		return nil, err
	}

	return GetCandlesticksRequest{
		Symbol:   req.Symbol,
		Interval: candlestick.Interval(req.Interval),
		From:     from,
		To:       to,
	}, err
}

func encodeGRPCGetCandlesticksResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(GetCandlesticksResponse)
	candlesticks := make(map[int64]*proto.Candlestick, len(resp.Candlesticks))
	for i := range resp.Candlesticks {
		candlesticks[int64(i)] = &proto.Candlestick{
			Open:      decimalToFloat32(resp.Candlesticks[i].Open),
			Low:       decimalToFloat32(resp.Candlesticks[i].Low),
			High:      decimalToFloat32(resp.Candlesticks[i].High),
			Close:     decimalToFloat32(resp.Candlesticks[i].Close),
			AdjClose:  decimalToFloat32(resp.Candlesticks[i].AdjClose),
			Volume:    int64(resp.Candlesticks[i].Volume),
			Timestamp: resp.Candlesticks[i].Timestamp.Unix(),
			Interval:  string(resp.Candlesticks[i].Interval),
			QuoteId:   resp.Candlesticks[i].QuoteId,
		}
	}
	return &proto.GetCandlesticksReply{Candlesticks: candlesticks, Err: err2str(resp.Err)}, nil
}

func err2str(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func decimalToFloat32(v decimal.Decimal) float32 {
	r, _ := v.Float64()
	return float32(r)
}
