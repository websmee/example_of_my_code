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
	"github.com/shopspring/decimal"
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
	fs := flag.NewFlagSet("quotes", flag.ExitOnError)
	var (
		quote          = fs.String("quote", "GC=F", "quote symbol")
		datasetName    = fs.String("dataset.name", "GC=F", "base name of the dataset")
		datasetPath    = fs.String("dataset.path", "./files/datasets/", "path to save datasets")
		trainFrom      = fs.String("dataset.trainFrom", "2020-01-01T00:00:00Z", "training dataset period")
		trainTo        = fs.String("dataset.trainTo", "2021-01-01T00:00:00Z", "training dataset period")
		testFrom       = fs.String("dataset.testFrom", "2021-01-01T00:00:00Z", "testing dataset period")
		testTo         = fs.String("dataset.testTo", "2021-02-01T00:00:00Z", "testing dataset period")
		normalizerPath = fs.String("normalizer.path", "./files/normalizers/", "path to save normalizers")
		paramsName     = fs.String("params.name", "CBS_GC=F", "name of the params")
		paramsPath     = fs.String("params.path", "./files/params/", "path to get/save params")
		quotesAddr     = fs.String("quotes.addr", "", "use this addr instead of consul discovery")
		consulAddr     = fs.String("consul.addr", "127.0.0.1", "consul address")
		consulPort     = fs.String("consul.port", "8500", "consul port")
		zipkinURL      = fs.String("zipkin-url", "", "Enable Zipkin tracing via HTTP reporter URL e.g. http://localhost:9411/api/v2/spans")
		zipkinBridge   = fs.Bool("zipkin-ot-bridge", false, "Use Zipkin OpenTracing bridge instead of native implementation")
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

	trainStart, _ := time.Parse(time.RFC3339, *trainFrom)
	trainEnd, _ := time.Parse(time.RFC3339, *trainTo)
	testStart, _ := time.Parse(time.RFC3339, *testFrom)
	testEnd, _ := time.Parse(time.RFC3339, *testTo)

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
		trainStart,
		testEnd,
	)
	datasetRepository := infrastructure.NewDatasetFileRepository(*datasetPath)
	normalizerRepository := infrastructure.NewNormalizerFileRepository(*normalizerPath)
	paramsRepository := infrastructure.NewParamsFileRepository(*paramsPath)
	var p []decimal.Decimal
	p, err = paramsRepository.LoadParams(*paramsName)
	if err != nil {
		_ = logger.Log("init", "cbsParams", "error", err, "stack", errors.GetStackTrace(err))
		return err
	}
	cbsParams := new(params.CBS)
	cbsParams.SetParams(p)
	datasetApp := app.NewCBSDatasetApp(candlestickCacheRepository, datasetRepository, normalizerRepository, cbsParams)

	// RUN

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		cancelFunc()
	}()

	if err := datasetApp.CreateDataset(ctx, *datasetName, *quote, trainStart, trainEnd, testStart, testEnd); err != nil {
		_ = logger.Log("run", "datasetApp", "error", err, "stack", errors.GetStackTrace(err))
		return err
	}

	_ = logger.Log("run", "exit")

	return nil
}
