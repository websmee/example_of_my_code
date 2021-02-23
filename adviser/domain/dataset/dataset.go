package dataset

import (
	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/nn"
)

type Dataset struct {
	Inputs  int
	Outputs int
	Debug   int
	Rows    []*Row
}

type Row struct {
	OriginalInputs   []decimal.Decimal
	NormalizedInputs []decimal.Decimal
	Outputs          []decimal.Decimal
	Debug            []decimal.Decimal
}

func NewCBSDataset() *Dataset {
	return &Dataset{
		Inputs:  nn.CBSInputs,
		Outputs: nn.CBSOutputs,
		Debug:   5,
	}
}

func (r Dataset) NewRow() *Row {
	return &Row{
		OriginalInputs:   make([]decimal.Decimal, 0, r.Inputs),
		NormalizedInputs: make([]decimal.Decimal, 0, r.Inputs),
		Outputs:          make([]decimal.Decimal, 0, r.Outputs),
		Debug:            make([]decimal.Decimal, 0, r.Debug),
	}
}

func (r *Row) NewInput(value decimal.Decimal) *Row {
	r.OriginalInputs = append(r.OriginalInputs, value)
	r.NormalizedInputs = append(r.NormalizedInputs, value)
	return r
}

func (r *Row) NewOutput(value decimal.Decimal) *Row {
	r.Outputs = append(r.Outputs, value)
	return r
}

func (r *Row) NewDebug(value decimal.Decimal) *Row {
	r.Debug = append(r.Debug, value)
	return r
}
