package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/openzipkin/zipkin-go"
	"github.com/websmee/ms/pkg/cmd"
	"github.com/websmee/ms/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/websmee/example_of_my_code/adviser/app"
	"github.com/websmee/example_of_my_code/adviser/cmd/dependencies"
	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
	"github.com/websmee/example_of_my_code/adviser/domain/params"
	"github.com/websmee/example_of_my_code/adviser/infrastructure"
	grpcInfra "github.com/websmee/example_of_my_code/adviser/infrastructure/grpc"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	fs := flag.NewFlagSet("test_params", flag.ExitOnError)
	var (
		paramsName   = fs.String("params.name", "CBS_test", "name of the params")
		paramsPath   = fs.String("params.path", "./files/params/", "path to get params")
		advicesPath  = fs.String("advices.path", "./files/advices/", "path to save results")
		periodFrom   = fs.String("tester.periodFrom", "2021-01-01T00:00:00Z", "testing params for this period")
		periodTo     = fs.String("tester.periodTo", "2021-04-01T00:00:00Z", "testing params for this period")
		quotesAddr   = fs.String("quotes.addr", "", "use this addr instead of consul discovery")
		consulAddr   = fs.String("consul.addr", "127.0.0.1", "consul address")
		consulPort   = fs.String("consul.port", "8500", "consul port")
		zipkinURL    = fs.String("zipkin-url", "", "Enable Zipkin tracing via HTTP reporter URL e.g. http://localhost:9411/api/v2/spans")
		zipkinBridge = fs.Bool("zipkin-ot-bridge", false, "Use Zipkin OpenTracing bridge instead of native implementation")
	)
	fs.Usage = cmd.UsageFor(fs, os.Args[0]+" [flags]")
	_ = fs.Parse(os.Args[1:])

	// DEPENDENCIES

	var (
		err          error
		logger       log.Logger
		zipkinTracer *zipkin.Tracer
		tracer       stdopentracing.Tracer
		quotesConn   *grpc.ClientConn
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

		quotesConn, err = dependencies.GetQuotesGRPCConnection(*quotesAddr, *consulAddr, *consulPort)
		if err != nil {
			_ = logger.Log("dependencies", "quotesConn", "error", err, "stack", errors.GetStackTrace(err))
			return err
		}
		defer quotesConn.Close()
	}

	// INIT

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	from, _ := time.Parse(time.RFC3339, *periodFrom)
	to, _ := time.Parse(time.RFC3339, *periodTo)

	var testerApp app.ParamsTesterApp
	{
		quotesApp := grpcInfra.NewQuotesAppGRPCClient(
			quotesConn,
			tracer,
			zipkinTracer,
			log.NewNopLogger(),
		)
		quoteRepository := infrastructure.NewQuoteGRPCRepository(quotesApp)
		candlestickCacheRepository, err := infrastructure.NewCandlestickCacheRepository(
			ctx,
			quoteRepository,
			infrastructure.NewCandlestickGRPCRepository(quotesApp),
			[]candlestick.Interval{candlestick.IntervalHour},
			from.Add(-params.TestOrderExpirationPeriod), // add extra period for current advice history
			to.Add(params.TestOrderExpirationPeriod),    // add extra period for expiration
		)
		if err != nil {
			_ = logger.Log("init", "candlestickCacheRepository", "error", err, "stack", errors.GetStackTrace(err))
			return err
		}
		testerApp = app.NewCBSScaledTesterApp(
			quoteRepository,
			candlestickCacheRepository,
			infrastructure.NewParamsFileRepository(*paramsPath),
			infrastructure.NewAdviceFileRepository(*advicesPath),
		)
	}

	// RUN

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		cancelFunc()
	}()

	if err := testerApp.TestParams(ctx, *paramsName, from, to); err != nil {
		_ = logger.Log("run", "testerApp", "error", err, "stack", errors.GetStackTrace(err))
		return err
	}

	_ = logger.Log("run", "exit")

	return nil
}
