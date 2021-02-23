package nn

import (
	"math"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat/distuv"
)

const (
	CBSInputs  = 1
	CBSHiddens = 1
	CBSOutputs = 3
)

type NN struct {
	Inputs        int
	Hiddens       int
	Outputs       int
	HiddenWeights *mat.Dense
	OutputWeights *mat.Dense
	LearningRate  float64
}

func NewCBSNN() *NN {
	return newNN(
		CBSInputs,
		CBSHiddens,
		CBSOutputs,
	)
}

func newNN(inputs, hiddens, outputs int) *NN {
	return &NN{
		Inputs:        inputs,
		Hiddens:       hiddens,
		Outputs:       outputs,
		HiddenWeights: mat.NewDense(hiddens, inputs, randomArray(inputs*hiddens, float64(inputs))),
		OutputWeights: mat.NewDense(outputs, hiddens, randomArray(hiddens*outputs, float64(hiddens))),
	}
}

func randomArray(size int, v float64) (data []float64) {
	dist := distuv.Uniform{
		Min: -1 / math.Sqrt(v),
		Max: 1 / math.Sqrt(v),
	}

	data = make([]float64, size)
	for i := 0; i < size; i++ {
		data[i] = dist.Rand()
	}
	return
}
