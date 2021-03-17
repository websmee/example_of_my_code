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
	"github.com/sony/gobreaker"
	"google.golang.org/grpc"

	"github.com/websmee/example_of_my_code/quotes/api/proto"

	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
	"github.com/websmee/example_of_my_code/adviser/domain/quote"
)

type QuotesApp interface {
	GetQuotes(ctx context.Context) ([]quote.Quote, error)
	GetCandlesticks(ctx context.Context, symbol string, interval candlestick.Interval, from, to time.Time) ([]candlestick.Candlestick, error)
}

type quotesAppGRPCClient struct {
	getQuotesEndpoint       endpoint.Endpoint
	getCandlesticksEndpoint endpoint.Endpoint
}

func NewQuotesAppGRPCClient(conn *grpc.ClientConn, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) QuotesApp {
	// limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 100))
	var options []grpctransport.ClientOption
	if zipkinTracer != nil {
		options = append(options, zipkin.GRPCClientTrace(zipkinTracer))
	}

	var getQuotesEndpoint endpoint.Endpoint
	{
		getQuotesEndpoint = grpctransport.NewClient(
			conn,
			"proto.Quotes",
			"GetQuotes",
			encodeGRPCGetQuotesRequest,
			decodeGRPCGetQuotesResponse,
			proto.GetQuotesReply{},
			append(options, grpctransport.ClientBefore(opentracing.ContextToGRPC(otTracer, logger)))...,
		).Endpoint()
		getQuotesEndpoint = opentracing.TraceClient(otTracer, "GetQuotes")(getQuotesEndpoint)
		// getQuotesEndpoint = limiter(getQuotesEndpoint)
		getQuotesEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "GetQuotes",
			Timeout: 30 * time.Second,
		}))(getQuotesEndpoint)
		getQuotesEndpoint = LoggingMiddleware(log.With(logger, "method", "GetQuotes"))(getQuotesEndpoint)
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

	return &quotesAppGRPCClient{
		getQuotesEndpoint:       getQuotesEndpoint,
		getCandlesticksEndpoint: getCandlesticksEndpoint,
	}
}
