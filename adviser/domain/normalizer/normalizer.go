package normalizer

import (
	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/dataset"
)

type Normalizer struct {
	InputsMin []decimal.Decimal
	InputsMax []decimal.Decimal
}

func NewDatasetNormalizer(ds *dataset.Dataset) *Normalizer {
	inputsMin := make([]decimal.Decimal, ds.Inputs)
	inputsMax := make([]decimal.Decimal, ds.Inputs)
	for i := range ds.Rows {
		for j := range ds.Rows[i].OriginalInputs {
			if inputsMin[j].GreaterThan(ds.Rows[i].OriginalInputs[j]) {
				inputsMin[j] = ds.Rows[i].OriginalInputs[j]
			}
			if inputsMax[j].LessThan(ds.Rows[i].OriginalInputs[j]) {
				inputsMax[j] = ds.Rows[i].OriginalInputs[j]
			}
		}
	}

	return &Normalizer{
		InputsMin: inputsMin,
		InputsMax: inputsMax,
	}
}

func (r Normalizer) NormalizeDataset(ds *dataset.Dataset) {
	if len(ds.Rows) == 0 {
		return
	}

	for i := range ds.Rows {
		for j := range ds.Rows[i].OriginalInputs {
			if r.InputsMax[j].Sub(r.InputsMin[j]).Equals(decimal.NewFromInt(0)) {
				ds.Rows[i].NormalizedInputs[j] = decimal.NewFromInt(1)
			} else {
				ds.Rows[i].NormalizedInputs[j] = ds.Rows[i].OriginalInputs[j].
					Sub(r.InputsMin[j]).Div(r.InputsMax[j].Sub(r.InputsMin[j]))
			}
		}
	}
}
