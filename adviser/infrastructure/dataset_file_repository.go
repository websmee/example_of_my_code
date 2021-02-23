package infrastructure

import (
	"encoding/csv"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/dataset"
)

type datasetFileRepository struct {
	filePath string
}

func NewDatasetFileRepository(filePath string) dataset.Repository {
	return &datasetFileRepository{
		filePath: filePath,
	}
}

func (r datasetFileRepository) SaveDataset(name string, ds *dataset.Dataset) error {
	f, err := os.Create(r.getFilepath(name))
	if err != nil {
		return errors.Wrap(err, "SaveDataset file create failed")
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	for i := range ds.Rows {
		record := make([]string, ds.Inputs+ds.Inputs+ds.Outputs+ds.Debug)
		for j := range ds.Rows[i].OriginalInputs {
			record[j] = ds.Rows[i].OriginalInputs[j].String()
		}
		for j := range ds.Rows[i].NormalizedInputs {
			record[ds.Inputs+j] = ds.Rows[i].NormalizedInputs[j].String()
		}
		for j := range ds.Rows[i].Outputs {
			record[ds.Inputs+ds.Inputs+j] = ds.Rows[i].Outputs[j].String()
		}
		for j := range ds.Rows[i].Debug {
			record[ds.Inputs+ds.Inputs+ds.Outputs+j] = ds.Rows[i].Debug[j].String()
		}

		if err := writer.Write(record); err != nil {
			return errors.Wrap(err, "SaveDataset file write failed")
		}
	}

	return nil
}

func (r datasetFileRepository) LoadDataset(name string, ds *dataset.Dataset) error {
	f, err := os.Open(r.getFilepath(name))
	if err != nil {
		return errors.Wrap(err, "LoadDataset file open failed")
	}
	defer f.Close()

	reader := csv.NewReader(f)
	i := 0
	for {
		record, err := reader.Read()
		if err == io.EOF || len(record) == 0 {
			break
		}

		ds.Rows = append(ds.Rows, ds.NewRow())
		for j := 0; j < ds.Inputs; j++ {
			x, _ := decimal.NewFromString(record[j])
			ds.Rows[i].OriginalInputs = append(ds.Rows[i].OriginalInputs, x)
		}
		for j := 0; j < ds.Inputs; j++ {
			x, _ := decimal.NewFromString(record[ds.Inputs+j])
			ds.Rows[i].NormalizedInputs = append(ds.Rows[i].NormalizedInputs, x)
		}
		for j := 0; j < ds.Outputs; j++ {
			x, _ := decimal.NewFromString(record[ds.Inputs+ds.Inputs+j])
			ds.Rows[i].Outputs = append(ds.Rows[i].Outputs, x)
		}
		for j := 0; j < ds.Debug; j++ {
			x, _ := decimal.NewFromString(record[ds.Inputs+ds.Inputs+ds.Outputs+j])
			ds.Rows[i].Debug = append(ds.Rows[i].Debug, x)
		}

		i++
	}

	return nil
}

func (r datasetFileRepository) getFilepath(name string) string {
	//todo: normalize filename
	return r.filePath + strings.ReplaceAll(name, "=", "_") + ".csv"
}
