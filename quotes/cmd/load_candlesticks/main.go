package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/go-kit/kit/log"
	"github.com/go-pg/pg/v9"

	"github.com/websmee/example_of_my_code/quotes/app"
	"github.com/websmee/example_of_my_code/quotes/infrastructure"
	"github.com/websmee/example_of_my_code/quotes/infrastructure/config"
	"github.com/websmee/example_of_my_code/quotes/infrastructure/persistence"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {

	// CMD INTERFACE

	fs := flag.NewFlagSet("quotes", flag.ExitOnError)
	var (
		consulAddr = fs.String("consul.addr", "127.0.0.1", "consul address")
		consulPort = fs.String("consul.port", "8500", "consul port")
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

	loader := app.NewCandlestickLoader(
		infrastructure.NewCandlestickYahooLoader(),
		persistence.NewCandlestickRepository(db),
		persistence.NewQuoteRepository(db),
	)

	// RUN

	if err := loader.LoadCandlesticks(); err != nil {
		_ = logger.Log("run", "nnApp", "error", fmt.Sprintf("%+v", err))
		return err
	}

	_ = logger.Log("run", "exit")

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
