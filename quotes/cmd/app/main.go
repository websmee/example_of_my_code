package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"

	"github.com/go-pg/pg/v9"

	"github.com/websmee/ms/pkg/discovery"
	"github.com/websmee/ms/pkg/discovery/health"
	healthProto "github.com/websmee/ms/pkg/discovery/health/proto"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/oklog/oklog/pkg/group"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	"github.com/websmee/example_of_my_code/quotes/api"
	"github.com/websmee/example_of_my_code/quotes/api/proto"
	"github.com/websmee/example_of_my_code/quotes/app"
	"github.com/websmee/example_of_my_code/quotes/infrastructure/config"
	"github.com/websmee/example_of_my_code/quotes/infrastructure/persistence"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {

	// CMD INTERFACE

	fs := flag.NewFlagSet("quotes", flag.ExitOnError)
	var (
		debugAddr         = fs.String("debug.addr", "0.0.0.0", "Debug and metrics listen address")
		debugPort         = fs.String("debug.port", "8080", "Debug and metrics listen port")
		grpcAddr          = fs.String("grpc.addr", "0.0.0.0", "gRPC listen address")
		grpcPort          = fs.String("grpc.port", "8082", "gRPC listen port")
		consulAddr        = fs.String("consul.addr", "127.0.0.1", "consul address")
		consulPort        = fs.String("consul.port", "8500", "consul port")
		consulServiceName = fs.String("consul.service_name", "quotes", "consul service name")
		consulServiceAddr = fs.String("consul.service_addr", "127.0.0.1", "consul service addr")
		consulServicePort = fs.String("consul.service_port", "8082", "consul service port")
		zipkinURL         = fs.String("zipkin-url", "", "Enable Zipkin tracing via HTTP reporter URL e.g. http://localhost:9411/api/v2/spans")
		zipkinBridge      = fs.Bool("zipkin-ot-bridge", false, "Use Zipkin OpenTracing bridge instead of native implementation")
	)
	fs.Usage = usageFor(fs, os.Args[0]+" [flags]")
	_ = fs.Parse(os.Args[1:])

	// LOGGER

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	// TRACER

	var zipkinTracer *zipkin.Tracer
	{
		if *zipkinURL != "" {
			var (
				err         error
				hostPort    = "localhost:80"
				serviceName = "quotes"
				reporter    = zipkinhttp.NewReporter(*zipkinURL)
			)
			defer reporter.Close()
			zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
			zipkinTracer, err = zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zEP))
			if err != nil {
				_ = logger.Log("tracer", "Zipkin", "error", fmt.Sprintf("%+v", err))
				return err
			}
			if !(*zipkinBridge) {
				_ = logger.Log("tracer", "Zipkin", "type", "Native", "URL", *zipkinURL)
			}
		}
	}

	var tracer stdopentracing.Tracer
	{
		if *zipkinBridge && zipkinTracer != nil {
			_ = logger.Log("tracer", "Zipkin", "type", "OpenTracing", "URL", *zipkinURL)
			tracer = zipkinot.Wrap(zipkinTracer)
			zipkinTracer = nil
		} else {
			tracer = stdopentracing.GlobalTracer() // no-op
		}
	}

	// METRICS

	var count metrics.Counter
	{
		// Business-level metrics.
		count = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "fintech",
			Subsystem: "quotes",
			Name:      "candlesticks_requested",
			Help:      "Total count of candlesticks requested.",
		}, []string{})
	}
	var duration metrics.Histogram
	{
		// Endpoint-level metrics.
		duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "fintech",
			Subsystem: "quotes",
			Name:      "request_duration_seconds",
			Help:      "Request duration in seconds.",
		}, []string{"method", "success"})
	}

	// CONFIG

	cfg, err := config.NewConsulKVConfig(*consulAddr+":"+*consulPort, logger)
	if err != nil {
		_ = logger.Log("config", "connect", "error", fmt.Sprintf("%+v", err))
		return err
	}

	dbConfig, err := cfg.GetDb("quotes_db")
	if err != nil {
		_ = logger.Log("config", "db", "error", fmt.Sprintf("%+v", err))
		return err
	}

	// DB

	db := pg.Connect(&pg.Options{
		Addr:     dbConfig.Host + ":" + dbConfig.Port,
		User:     dbConfig.User,
		Password: dbConfig.Password,
		Database: dbConfig.Name,
	})
	defer db.Close()

	err = persistence.Migrate(db)
	if err != nil {
		_ = logger.Log("db", "migrate", "error", fmt.Sprintf("%+v", err))
		return err
	}

	// INIT

	var (
		extLogger       = log.NewNopLogger() // change it to logger to see extended log in console
		quoteRepo       = persistence.NewQuoteRepository(db)
		candlestickRepo = persistence.NewCandlestickRepository(db)
		quotes          = app.NewQuotesApp(extLogger, count, quoteRepo, candlestickRepo)
		endpoints       = api.NewQuotes(quotes, extLogger, duration, tracer, zipkinTracer)
		grpcServer      = api.NewGRPCServer(endpoints, tracer, zipkinTracer, logger)

		healthCheckEndpoint = health.NewCheckEndpoint(func(service string) health.CheckStatus {
			if quotes.HealthCheck() {
				return health.CheckStatusServing
			}
			return health.CheckStatusNotServing
		}, extLogger, duration, tracer, zipkinTracer)
		grpcHealthServer = health.NewGRPCServer(healthCheckEndpoint, tracer, zipkinTracer, logger)

		serviceRegistrar = discovery.NewConsulRegistrar(*consulAddr+":"+*consulPort, logger)
	)

	var g group.Group
	{
		// HTTP

		http.DefaultServeMux.Handle("/metrics", promhttp.Handler())

		addr := *debugAddr + ":" + *debugPort
		debugListener, err := net.Listen("tcp", addr)
		if err != nil {
			_ = logger.Log("transport", "debug/HTTP", "during", "Listen", "error", fmt.Sprintf("%+v", err))
			return err
		}
		g.Add(func() error {
			_ = logger.Log("transport", "debug/HTTP", "addr", addr)
			return http.Serve(debugListener, http.DefaultServeMux)
		}, func(error) {
			debugListener.Close()
		})
	}
	{
		// GRPC

		addr := *grpcAddr + ":" + *grpcPort
		grpcListener, err := net.Listen("tcp", addr)
		if err != nil {
			_ = logger.Log("transport", "gRPC", "during", "Listen", "error", fmt.Sprintf("%+v", err))
			return err
		}
		g.Add(func() error {
			// register service in consul
			if err := serviceRegistrar.Register(*consulServiceName, *consulServiceAddr+":"+*consulServicePort); err != nil {
				_ = logger.Log("transport", "gRPC", "during", "Register", "error", fmt.Sprintf("%+v", err))
				return err
			}

			// serve
			_ = logger.Log("transport", "gRPC", "addr", addr)
			baseServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
			proto.RegisterQuotesServer(baseServer, grpcServer)
			healthProto.RegisterHealthServer(baseServer, grpcHealthServer)
			return baseServer.Serve(grpcListener)
		}, func(error) {
			grpcListener.Close()
			serviceRegistrar.DeregisterAll()
		})
	}
	{
		// This function just sits and waits for ctrl-C.
		cancelInterrupt := make(chan struct{})
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}
	_ = logger.Log("exit", g.Run())

	return nil
}

func usageFor(fs *flag.FlagSet, short string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s\n", short)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		w.Flush()
		fmt.Fprintf(os.Stderr, "\n")
	}
}
