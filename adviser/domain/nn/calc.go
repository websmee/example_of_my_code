package nn

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

type Calculator interface {
	Dot(m, n mat.Matrix) mat.Matrix
	Apply(fn func(i, j int, v float64) float64, m mat.Matrix) mat.Matrix
	Scale(s float64, m mat.Matrix) mat.Matrix
	Multiply(m, n mat.Matrix) mat.Matrix
	Add(m, n mat.Matrix) mat.Matrix
	Subtract(m, n mat.Matrix) mat.Matrix
	Sigmoid(a, c int, z float64) float64
	SigmoidPrime(m mat.Matrix) mat.Matrix
}

type calculator struct {
}

func DefaultCalculator() Calculator {
	return &calculator{}
}

func (r calculator) Dot(m, n mat.Matrix) mat.Matrix {
	a, _ := m.Dims()
	_, c := n.Dims()
	o := mat.NewDense(a, c, nil)
	o.Product(m, n)
	return o
}

func (r calculator) Apply(fn func(i, j int, v float64) float64, m mat.Matrix) mat.Matrix {
	a, c := m.Dims()
	o := mat.NewDense(a, c, nil)
	o.Apply(fn, m)
	return o
}

func (r calculator) Scale(s float64, m mat.Matrix) mat.Matrix {
	a, c := m.Dims()
	o := mat.NewDense(a, c, nil)
	o.Scale(s, m)
	return o
}

func (r calculator) Multiply(m, n mat.Matrix) mat.Matrix {
	a, c := m.Dims()
	o := mat.NewDense(a, c, nil)
	o.MulElem(m, n)
	return o
}

func (r calculator) Add(m, n mat.Matrix) mat.Matrix {
	a, c := m.Dims()
	o := mat.NewDense(a, c, nil)
	o.Add(m, n)
	return o
}

func (r calculator) Subtract(m, n mat.Matrix) mat.Matrix {
	a, c := m.Dims()
	o := mat.NewDense(a, c, nil)
	o.Sub(m, n)
	return o
}

func (r calculator) Sigmoid(_, _ int, z float64) float64 {
	return 1.0 / (1 + math.Exp(-1*z))
}

func (r calculator) SigmoidPrime(m mat.Matrix) mat.Matrix {
	rows, _ := m.Dims()
	o := make([]float64, rows)
	for i := range o {
		o[i] = 1
	}
	ones := mat.NewDense(rows, 1, o)
	return r.Multiply(m, r.Subtract(ones, m)) // m * (1 - m)
}
