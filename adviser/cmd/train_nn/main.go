package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/websmee/ms/pkg/cmd"
	"github.com/websmee/ms/pkg/errors"
	"golang.org/x/net/context"

	"github.com/websmee/example_of_my_code/adviser/app"
	"github.com/websmee/example_of_my_code/adviser/cmd/dependencies"
	"github.com/websmee/example_of_my_code/adviser/infrastructure"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	fs := flag.NewFlagSet("quotes", flag.ExitOnError)
	var (
		nnName       = fs.String("nn.name", "GC=F_1000_001", "name of the nn")
		nnPath       = fs.String("nn.path", "./files/nn_models/", "path to get/save nn models")
		datasetName  = fs.String("dataset.name", "GC=F-train", "name of the dataset")
		datasetPath  = fs.String("dataset.path", "./files/datasets/", "path to get datasets")
		epochs       = fs.Int("train.epochs", 1000, "number of training epochs")
		learningRate = fs.Float64("train.rate", 0.01, "learning rate")
	)
	fs.Usage = cmd.UsageFor(fs, os.Args[0]+" [flags]")
	_ = fs.Parse(os.Args[1:])

	// DEPENDENCIES

	logger := dependencies.GetLogger()

	// INIT

	nnRepository := infrastructure.NewNNFileRepository(*nnPath)
	datasetRepository := infrastructure.NewDatasetFileRepository(*datasetPath)
	nnTrainerApp := app.NewNNTrainerApp(nnRepository, datasetRepository)

	// RUN

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		cancelFunc()
	}()

	if err := nnTrainerApp.TrainNN(ctx, *nnName, *datasetName, *epochs, *learningRate); err != nil {
		_ = logger.Log("run", "nnTrainerApp", "error", err, "stack", errors.GetStackTrace(err))
		return err
	}

	_ = logger.Log("run", "exit")

	return nil
}
