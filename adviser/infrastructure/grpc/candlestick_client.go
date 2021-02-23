package grpc

import (
	"context"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/shopspring/decimal"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc"

	"github.com/websmee/example_of_my_code/quotes/api/proto"

	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
)

func NewCandlestickGRPCClient(conn *grpc.ClientConn, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) candlestick.Repository {
	// limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 100))
	var options []grpctransport.ClientOption
	if zipkinTracer != nil {
		options = append(options, zipkin.GRPCClientTrace(zipkinTracer))
	}

	var getCandlesticksEndpoint endpoint.Endpoint
	{
		getCandlesticksEndpoint = grpctransport.NewClient(
			conn,
			"proto.Quotes",
			"GetCandlesticks",
			encodeGRPCGetCandlesticksRequest,
			decodeGRPCGetCandlesticksResponse,
			proto.GetCandlesticksReply{},
			append(options, grpctransport.ClientBefore(opentracing.ContextToGRPC(otTracer, logger)))...,
		).Endpoint()
		getCandlesticksEndpoint = opentracing.TraceClient(otTracer, "GetCandlesticks")(getCandlesticksEndpoint)
		// getCandlesticksEndpoint = limiter(getCandlesticksEndpoint)
		getCandlesticksEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "GetCandlesticks",
			Timeout: 30 * time.Second,
		}))(getCandlesticksEndpoint)
		getCandlesticksEndpoint = LoggingMiddleware(log.With(logger, "method", "GetCandlesticks"))(getCandlesticksEndpoint)
	}

	return CandlestickRepository{
		GetCandlesticksEndpoint: getCandlesticksEndpoint,
	}
}

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
			QuoteId:   reply.Candlesticks[i].QuoteId,
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
