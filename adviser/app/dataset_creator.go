package app

import (
	"context"
	"time"

	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
	"github.com/websmee/example_of_my_code/adviser/domain/dataset"
	"github.com/websmee/example_of_my_code/adviser/domain/normalizer"
	"github.com/websmee/example_of_my_code/adviser/domain/params"
)

type DatasetCreatorApp interface {
	CreateDataset(ctx context.Context, datasetName, quoteSymbol string, trainStart, trainEnd, testStart, testEnd time.Time) error
}

type cbsDatasetCreatorApp struct {
	factory              dataset.Factory
	datasetRepository    dataset.Repository
	normalizerRepository normalizer.Repository
	cbsParams            *params.CBS
}

func NewCBSDatasetApp(
	candlestickRepository candlestick.Repository,
	datasetRepository dataset.Repository,
	normalizerRepository normalizer.Repository,
	cbsParams *params.CBS,
) DatasetCreatorApp {
	return &cbsDatasetCreatorApp{
		factory: dataset.NewFactory(
			candlestick.NewBasicFilter(candlestickRepository),
		),
		datasetRepository:    datasetRepository,
		normalizerRepository: normalizerRepository,
		cbsParams:            cbsParams,
	}
}

func (r cbsDatasetCreatorApp) CreateDataset(
	ctx context.Context,
	datasetName,
	quoteSymbol string,
	trainStart,
	trainEnd,
	testStart,
	testEnd time.Time,
) error {
	n, err := r.prepareTrain(ctx, datasetName, quoteSymbol, trainStart, trainEnd)
	if err != nil {
		return err
	}
	if err := r.prepareTest(ctx, datasetName, quoteSymbol, testStart, testEnd, n); err != nil {
		return err
	}
	if err := r.normalizerRepository.SaveNormalizer(quoteSymbol, n); err != nil {
		return err
	}

	return nil
}

func (r cbsDatasetCreatorApp) prepareTrain(
	ctx context.Context,
	datasetName,
	quoteSymbol string,
	start time.Time,
	end time.Time,
) (*normalizer.Normalizer, error) {
	ds, err := r.factory.CreateCBSDataset(ctx, r.cbsParams, quoteSymbol, start, end)
	if err != nil {
		return nil, err
	}

	n := normalizer.NewDatasetNormalizer(ds)
	n.NormalizeDataset(ds)

	if err := r.datasetRepository.SaveDataset(datasetName, ds); err != nil {
		return nil, err
	}

	return n, nil
}

func (r cbsDatasetCreatorApp) prepareTest(
	ctx context.Context,
	datasetName,
	quoteSymbol string,
	start time.Time,
	end time.Time,
	n *normalizer.Normalizer,
) error {
	ds, err := r.factory.CreateCBSDataset(ctx, r.cbsParams, quoteSymbol, start, end)
	if err != nil {
		return err
	}

	n.NormalizeDataset(ds)

	if err := r.datasetRepository.SaveDataset(datasetName, ds); err != nil {
		return err
	}

	return nil
}
