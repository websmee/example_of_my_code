package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/websmee/ms/pkg/cmd"
	"github.com/websmee/ms/pkg/errors"

	"github.com/websmee/example_of_my_code/quotes/cmd/dependencies"
	"github.com/websmee/example_of_my_code/quotes/infrastructure"
	"github.com/websmee/example_of_my_code/quotes/infrastructure/tiingo"

	"github.com/websmee/ms/pkg/discovery"
	"github.com/websmee/ms/pkg/discovery/health"
	healthProto "github.com/websmee/ms/pkg/discovery/health/proto"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/oklog/oklog/pkg/group"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/openzipkin/zipkin-go"
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
		dbMigrationsPath  = fs.String("db-migrations-path", "infrastructure/persistence/migrations/", "Where to find migrations")
	)
	fs.Usage = cmd.UsageFor(fs, os.Args[0]+" [flags]")
	_ = fs.Parse(os.Args[1:])

	// DEPENDENCIES

	var (
		err          error
		logger       log.Logger
		zipkinTracer *zipkin.Tracer
		tracer       stdopentracing.Tracer
		onclose      func()
	)
	{
		logger = dependencies.GetLogger()
		zipkinTracer, tracer, onclose, err = dependencies.GetTracers(*zipkinURL, *zipkinBridge)
		if err != nil {
			_ = logger.Log("dependencies", "tracer", "error", err, "stack", errors.GetStackTrace(err))
			return err
		}
		defer onclose()
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

	var (
		dbConfig     *config.DB
		tiingoConfig *config.Tiingo
	)
	{
		cfg, err := config.NewConsulKVConfig(*consulAddr+":"+*consulPort, logger)
		if err != nil {
			_ = logger.Log("config", "connect", "error", err, "stack", errors.GetStackTrace(err))
			return err
		}

		dbConfig, err = cfg.GetDB("quotes_db")
		if err != nil {
			_ = logger.Log("config", "db", "error", err, "stack", errors.GetStackTrace(err))
			return err
		}

		tiingoConfig, err = cfg.GetTiingo()
		if err != nil {
			_ = logger.Log("config", "tiingo", "error", err, "stack", errors.GetStackTrace(err))
			return err
		}
	}

	// DB

	var db *pg.DB
	{
		db = pg.Connect(&pg.Options{
			Addr:     dbConfig.Host + ":" + dbConfig.Port,
			User:     dbConfig.User,
			Password: dbConfig.Password,
			Database: dbConfig.Name,
		})
		defer db.Close()

		if err := persistence.Migrate(db, *dbMigrationsPath); err != nil {
			_ = logger.Log("db", "migrate", "error", err, "stack", errors.GetStackTrace(err))
			return err
		}
	}

	// INIT

	var (
		extLogger       = logger
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

		candlestickLoader = app.NewCandlestickLoader(
			extLogger,
			infrastructure.NewTiingoCandlestickLoader(tiingo.NewClient(tiingoConfig)),
			candlestickRepo,
			quoteRepo,
		)
	)

	var g group.Group
	{
		// HTTP

		http.DefaultServeMux.Handle("/metrics", promhttp.Handler())

		addr := *debugAddr + ":" + *debugPort
		debugListener, err := net.Listen("tcp", addr)
		if err != nil {
			_ = logger.Log("transport", "debug/HTTP", "during", "Listen", "error", err, "stack", errors.GetStackTrace(err))
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
			_ = logger.Log("transport", "gRPC", "during", "Listen", "error", err, "stack", errors.GetStackTrace(err))
			return err
		}
		g.Add(func() error {
			// register service in consul
			if err := serviceRegistrar.Register(*consulServiceName, *consulServiceAddr+":"+*consulServicePort); err != nil {
				_ = logger.Log("transport", "gRPC", "during", "Register", "error", err, "stack", errors.GetStackTrace(err))
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
		// LOAD LATEST CANDLESTICKS

		ticker := time.NewTicker(time.Minute)
		g.Add(func() error {
			for range ticker.C {
				if time.Now().Minute() == 1 { // every first minute of every hour
					if err := candlestickLoader.LoadLatest(); err != nil {
						_ = logger.Log("background task", "load latest candlesticks", "error", err, "stack", errors.GetStackTrace(err))
					}
				}
			}

			return nil
		}, func(error) {
			ticker.Stop()
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
