package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/openzipkin/zipkin-go"
	"github.com/websmee/ms/pkg/cmd"
	"github.com/websmee/ms/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/websmee/example_of_my_code/adviser/app"
	"github.com/websmee/example_of_my_code/adviser/cmd/dependencies"
	"github.com/websmee/example_of_my_code/adviser/infrastructure"
	grpcInfra "github.com/websmee/example_of_my_code/adviser/infrastructure/grpc"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	fs := flag.NewFlagSet("advice_viewer", flag.ExitOnError)
	var (
		addr         = fs.String("http.addr", "", "viewer listen address")
		port         = fs.String("http.port", "80", "viewer listen port")
		advicesPath  = fs.String("advices.path", "./files/advices/", "path to get/save advices")
		advicesName  = fs.String("advices.name", "CBS_test_loss", "name of the set of advices")
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

	quotesApp := grpcInfra.NewQuotesAppGRPCClient(
		quotesConn,
		tracer,
		zipkinTracer,
		log.NewNopLogger(),
	)
	candlestickRepository := infrastructure.NewCandlestickGRPCRepository(quotesApp)
	adviceRepository := infrastructure.NewAdviceFileRepository(*advicesPath)
	viewerApp := app.NewViewerApp(adviceRepository, candlestickRepository)

	// RUN

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		cancelFunc()
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		charts, err := viewerApp.GetCharts(ctx, *advicesName, 0, 1)
		if err != nil {
			panic(err)
		}

		for i := range charts {
			err = charts[i].Render(w)
			if err != nil {
				panic(err)
			}
		}
	})
	_ = http.ListenAndServe(*addr+":"+*port, nil)

	_ = logger.Log("run", "exit")

	return nil
}
