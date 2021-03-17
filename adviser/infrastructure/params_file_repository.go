package infrastructure

import (
	"encoding/csv"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/params"
)

type paramsFileRepository struct {
	filePath string
}

func NewParamsFileRepository(filePath string) params.Repository {
	return &paramsFileRepository{
		filePath: filePath,
	}
}

func (r paramsFileRepository) SaveParams(name string, params []decimal.Decimal) error {
	f, err := os.Create(r.getFilepath(name))
	if err != nil {
		return errors.Wrap(err, "SaveSignal file create failed")
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	record := make([]string, len(params))
	for i, p := range params {
		record[i] = p.String()
	}
	if err := writer.Write(record); err != nil {
		return errors.Wrap(err, "SaveSignal file write failed")
	}

	return nil
}

func (r paramsFileRepository) LoadParams(name string) ([]decimal.Decimal, error) {
	f, err := os.Open(r.getFilepath(name))
	if err != nil {
		return nil, errors.Wrap(err, "LoadSignal file open failed")
	}
	defer f.Close()

	{
		reader := csv.NewReader(f)
		record, err := reader.Read()
		if err == io.EOF || len(record) == 0 {
			return nil, errors.Wrap(err, "LoadSignal file read failed")
		}

		var ps []decimal.Decimal
		for i := range record {
			p, err := decimal.NewFromString(record[i])
			if err != nil {
				return nil, errors.Wrap(err, "LoadSignal param parsing failed")
			}
			ps = append(ps, p)
		}

		return ps, nil
	}
}

func (r paramsFileRepository) getFilepath(name string) string {
	//todo: normalize filename
	return r.filePath + strings.ReplaceAll(name, "=", "_") + ".csv"
}
