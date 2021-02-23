package api

import (
	"context"
	"time"

	"github.com/websmee/example_of_my_code/quotes/app"
	"github.com/websmee/example_of_my_code/quotes/domain/candlestick"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
)

type Quotes struct {
	GetCandlesticksEndpoint endpoint.Endpoint
}

func NewQuotes(svc app.QuotesApp, logger log.Logger, duration metrics.Histogram, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) Quotes {
	var getCandlesticksEndpoint endpoint.Endpoint
	{
		getCandlesticksEndpoint = MakeGetCandlesticksEndpoint(svc)
		// getCandlesticksEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(getCandlesticksEndpoint)
		// getCandlesticksEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(getCandlesticksEndpoint)
		getCandlesticksEndpoint = opentracing.TraceServer(otTracer, "GetCandlesticks")(getCandlesticksEndpoint)
		if zipkinTracer != nil {
			getCandlesticksEndpoint = zipkin.TraceEndpoint(zipkinTracer, "GetCandlesticks")(getCandlesticksEndpoint)
		}
		getCandlesticksEndpoint = LoggingMiddleware(log.With(logger, "method", "GetCandlesticks"))(getCandlesticksEndpoint)
		getCandlesticksEndpoint = InstrumentingMiddleware(duration.With("method", "GetCandlesticks"))(getCandlesticksEndpoint)
	}
	return Quotes{
		GetCandlesticksEndpoint: getCandlesticksEndpoint,
	}
}

func MakeGetCandlesticksEndpoint(s app.QuotesApp) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetCandlesticksRequest)
		candlesticks, err := s.GetCandlesticks(req.Symbol, req.Interval, req.From, req.To)
		return GetCandlesticksResponse{Candlesticks: candlesticks, Err: err}, nil
	}
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
