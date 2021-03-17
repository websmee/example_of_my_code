package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-pg/pg/v9"
	"github.com/websmee/ms/pkg/cmd"
	"github.com/websmee/ms/pkg/errors"

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
	fs := flag.NewFlagSet("quotes", flag.ExitOnError)
	var (
		consulAddr       = fs.String("consul.addr", "127.0.0.1", "consul address")
		consulPort       = fs.String("consul.port", "8500", "consul port")
		dbMigrationsPath = fs.String("db-migrations-path", "infrastructure/persistence/migrations/", "Where to find migrations")
	)
	fs.Usage = cmd.UsageFor(fs, os.Args[0]+" [flags]")
	_ = fs.Parse(os.Args[1:])

	// LOGGER

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	// CONFIG

	var (
		dbConfig           *config.DB
		alphaVantageConfig *config.AlphaVantage
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

		alphaVantageConfig, err = cfg.GetAlphaVantage()
		if err != nil {
			_ = logger.Log("config", "apikey", "error", err, "stack", errors.GetStackTrace(err))
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

	loader := app.NewCandlestickLoader(
		infrastructure.NewCandlestickAlphaVantageLoader(alphaVantageConfig),
		persistence.NewCandlestickRepository(db),
		persistence.NewQuoteRepository(db),
	)

	// RUN

	if err := loader.LoadCandlesticks(); err != nil {
		_ = logger.Log("run", "nnApp", "error", err, "stack", errors.GetStackTrace(err))
		return err
	}

	_ = logger.Log("run", "exit")

	return nil
}
