package nn

import (
	"gonum.org/v1/gonum/mat"
)

type Trainer interface {
	Train(net *NN, inputData []float64, targetData []float64, learningRate float64)
}

type trainer struct {
	calc Calculator
}

func DefaultTrainer(calc Calculator) Trainer {
	return &trainer{calc}
}

func (r trainer) Train(net *NN, inputData []float64, targetData []float64, learningRate float64) {
	// forward propagation
	inputs := mat.NewDense(len(inputData), 1, inputData)
	hiddenInputs := r.calc.Dot(net.HiddenWeights, inputs)
	hiddenOutputs := r.calc.Apply(r.calc.Sigmoid, hiddenInputs)
	finalInputs := r.calc.Dot(net.OutputWeights, hiddenOutputs)
	finalOutputs := r.calc.Apply(r.calc.Sigmoid, finalInputs)

	// find errors
	targets := mat.NewDense(len(targetData), 1, targetData)
	outputErrors := r.calc.Subtract(targets, finalOutputs)
	hiddenErrors := r.calc.Dot(net.OutputWeights.T(), outputErrors)

	// backpropagate
	net.OutputWeights = r.calc.Add(net.OutputWeights,
		r.calc.Scale(learningRate,
			r.calc.Dot(r.calc.Multiply(outputErrors, r.calc.SigmoidPrime(finalOutputs)),
				hiddenOutputs.T()))).(*mat.Dense)

	net.HiddenWeights = r.calc.Add(net.HiddenWeights,
		r.calc.Scale(learningRate,
			r.calc.Dot(r.calc.Multiply(hiddenErrors, r.calc.SigmoidPrime(hiddenOutputs)),
				inputs.T()))).(*mat.Dense)
}
