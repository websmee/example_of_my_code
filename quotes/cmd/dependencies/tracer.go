package dependencies

import (
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

func GetTracers(zipkinURL string, zipkinBridge bool) (*zipkin.Tracer, stdopentracing.Tracer, func(), error) {
	var (
		zipkinTracer *zipkin.Tracer
		onclose      func()
	)
	{
		if zipkinURL != "" {
			var (
				err         error
				hostPort    = "localhost:80"
				serviceName = "quotes"
				reporter    = zipkinhttp.NewReporter(zipkinURL)
			)
			onclose = func() {
				reporter.Close()
			}
			zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
			zipkinTracer, err = zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zEP))
			if err != nil {
				reporter.Close()
				return nil, nil, nil, err
			}
		}
	}

	var tracer stdopentracing.Tracer
	{
		if zipkinBridge && zipkinTracer != nil {
			tracer = zipkinot.Wrap(zipkinTracer)
			zipkinTracer = nil
		} else {
			tracer = stdopentracing.GlobalTracer() // no-op
		}
	}

	return zipkinTracer, tracer, onclose, nil
}
