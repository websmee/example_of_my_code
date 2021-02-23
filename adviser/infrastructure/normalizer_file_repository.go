package infrastructure

import (
	"encoding/csv"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/normalizer"
)

type normalizerFileRepository struct {
	filePath string
}

func NewNormalizerFileRepository(filePath string) normalizer.Repository {
	return &normalizerFileRepository{
		filePath: filePath,
	}
}

func (r normalizerFileRepository) SaveNormalizer(name string, n *normalizer.Normalizer) error {
	f, err := os.Create(r.getFilepath(name))
	if err != nil {
		return errors.Wrap(err, "SaveNormalizer file create failed")
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	recordMin := make([]string, len(n.InputsMin))
	for i := range n.InputsMin {
		recordMin[i] = n.InputsMin[i].String()
	}
	if err := writer.Write(recordMin); err != nil {
		return errors.Wrap(err, "SaveNormalizer file write failed")
	}

	recordMax := make([]string, len(n.InputsMax))
	for i := range n.InputsMax {
		recordMax[i] = n.InputsMax[i].String()
	}
	if err := writer.Write(recordMax); err != nil {
		return errors.Wrap(err, "SaveNormalizer file write failed")
	}

	return nil
}

func (r normalizerFileRepository) LoadNormalizer(name string) (*normalizer.Normalizer, error) {
	f, err := os.Open(r.getFilepath(name))
	if err != nil {
		return nil, errors.Wrap(err, "LoadNormalizer file open failed")
	}
	defer f.Close()

	reader := csv.NewReader(f)
	n := &normalizer.Normalizer{}
	i := 0
	for {
		record, err := reader.Read()
		if err == io.EOF || len(record) == 0 {
			return nil, errors.Wrap(err, "LoadNormalizer file read failed")
		}

		if i == 0 {
			for j := range record {
				n.InputsMin[i], _ = decimal.NewFromString(record[j])
			}
		}

		if i == 1 {
			for j := range record {
				n.InputsMax[i], _ = decimal.NewFromString(record[j])
			}
			break
		}

		i++
	}

	return n, nil
}

func (r normalizerFileRepository) getFilepath(name string) string {
	//todo: normalize filename
	return r.filePath + strings.ReplaceAll(name, "=", "_") + ".csv"
}
