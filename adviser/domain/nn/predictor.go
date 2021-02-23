package nn

import (
	"gonum.org/v1/gonum/mat"
)

type Predictor interface {
	Predict(net *NN, inputData []float64) mat.Matrix
}

type predictor struct {
	calc Calculator
}

func DefaultPredictor(calc Calculator) Predictor {
	return &predictor{calc}
}

func (r predictor) Predict(net *NN, inputData []float64) mat.Matrix {
	// forward propagation
	inputs := mat.NewDense(len(inputData), 1, inputData)
	hiddenInputs := r.calc.Dot(net.HiddenWeights, inputs)
	hiddenOutputs := r.calc.Apply(r.calc.Sigmoid, hiddenInputs)
	finalInputs := r.calc.Dot(net.OutputWeights, hiddenOutputs)
	finalOutputs := r.calc.Apply(r.calc.Sigmoid, finalInputs)
	return finalOutputs
}
