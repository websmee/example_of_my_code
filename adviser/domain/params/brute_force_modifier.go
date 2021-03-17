package params

import (
	"github.com/shopspring/decimal"
)

type bruteForceModifier struct {
	step         int
	started      bool
	rate         decimal.Decimal
	currentParam int
	min          []decimal.Decimal
	max          []decimal.Decimal
}

func NewBruteForceParamsModifier(min, max []decimal.Decimal, rate decimal.Decimal) Modifier {
	return &bruteForceModifier{
		step:         0,
		started:      false,
		rate:         rate,
		currentParam: 0,
		min:          min,
		max:          max,
	}
}

func (r *bruteForceModifier) Modify(modifying []decimal.Decimal) (modified bool) {
	if !r.started {
		r.step = 0
		r.reset(modifying, 0)
		r.started = true
		return true
	}

	r.current()

	for modifying[r.currentParam].GreaterThanOrEqual(r.max[r.currentParam]) {
		if r.currentParam == 0 {
			r.started = false
			return false
		}
		r.prev()
	}

	for r.currentParam < len(modifying)-1 &&
		modifying[r.currentParam+1].Equals(r.min[r.currentParam+1]) &&
		!modifying[r.currentParam+1].Equals(r.max[r.currentParam+1]) {
		r.next()
	}

	r.step++
	modifying[r.currentParam] = modifying[r.currentParam].Add(r.getRate())
	r.reset(modifying, r.currentParam+1)
	if r.currentParam < len(modifying)-1 {
		r.next()
	}

	return true
}

func (r *bruteForceModifier) GetCurrentStep() int {
	return r.step
}

func (r *bruteForceModifier) GetTotalSteps() int {
	pow := decimal.NewFromInt(int64(len(r.min)))
	for i := range r.min {
		if r.min[i].Equals(r.max[i]) {
			pow = pow.Sub(decimal.NewFromInt(1))
		}
	}

	return int(decimal.NewFromInt(1).Div(r.rate).Ceil().Add(decimal.NewFromInt(1)).Pow(pow).IntPart())
}

func (r *bruteForceModifier) reset(modifying []decimal.Decimal, offset int) {
	for i := offset; i < len(modifying); i++ {
		modifying[i] = r.min[i]
	}
}

func (r *bruteForceModifier) getRate() decimal.Decimal {
	return r.max[r.currentParam].Sub(r.min[r.currentParam]).Mul(r.rate)
}

func (r *bruteForceModifier) current() {
	for r.min[r.currentParam].Equals(r.max[r.currentParam]) {
		r.currentParam++
		if r.currentParam == len(r.min) {
			r.currentParam = 0
			break
		}
	}
}

func (r *bruteForceModifier) next() {
	start := r.currentParam
	r.currentParam++
	for r.min[r.currentParam].Equals(r.max[r.currentParam]) {
		r.currentParam++
		if r.currentParam == len(r.min) {
			r.currentParam = start
			break
		}
	}
}

func (r *bruteForceModifier) prev() {
	r.currentParam--
	for r.min[r.currentParam].Equals(r.max[r.currentParam]) {
		r.currentParam--
		if r.currentParam < 0 {
			r.currentParam = 0
			break
		}
	}
}
