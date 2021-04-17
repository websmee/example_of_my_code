package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"

	"github.com/websmee/example_of_my_code/adviser/app"
	"github.com/websmee/example_of_my_code/adviser/domain/advice"
)

type Adviser struct {
	GetAdvicesEndpoint endpoint.Endpoint
}

func NewAdviser(svc app.AdviserApp, logger log.Logger, duration metrics.Histogram, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) Adviser {
	var getAdvicesEndpoint endpoint.Endpoint
	{
		getAdvicesEndpoint = MakeGetAdvicesEndpoint(svc)
		// getAdvicesEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(getAdvicesEndpoint)
		// getAdvicesEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(getAdvicesEndpoint)
		getAdvicesEndpoint = opentracing.TraceServer(otTracer, "GetAdvices")(getAdvicesEndpoint)
		if zipkinTracer != nil {
			getAdvicesEndpoint = zipkin.TraceEndpoint(zipkinTracer, "GetAdvices")(getAdvicesEndpoint)
		}
		getAdvicesEndpoint = LoggingMiddleware(log.With(logger, "method", "GetAdvices"))(getAdvicesEndpoint)
		getAdvicesEndpoint = InstrumentingMiddleware(duration.With("method", "GetAdvices"))(getAdvicesEndpoint)
	}
	return Adviser{
		GetAdvicesEndpoint: getAdvicesEndpoint,
	}
}

func MakeGetAdvicesEndpoint(s app.AdviserApp) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		advices, err := s.GetAdvices(ctx)
		return GetAdvicesResponse{Advices: advices, Err: err}, nil
	}
}

var (
	_ endpoint.Failer = GetAdvicesResponse{}
)

type GetAdvicesRequest struct{}

type GetAdvicesResponse struct {
	Advices []advice.Advice
	Err     error
}

func (r GetAdvicesResponse) Failed() error { return r.Err }
