package params

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestBruteForceModifier_Modify(t *testing.T) {
	min := []decimal.Decimal{
		decimal.NewFromInt(10),
		decimal.NewFromInt(20),
		decimal.NewFromInt(30),
	}
	max := []decimal.Decimal{
		decimal.NewFromInt(20),
		decimal.NewFromInt(40),
		decimal.NewFromInt(30),
	}
	current := []decimal.Decimal{
		decimal.NewFromInt(11),
		decimal.NewFromInt(22),
		decimal.NewFromInt(33),
	}
	rate := decimal.NewFromFloat(0.2)

	m := NewBruteForceParamsModifier(min, max, rate)
	i := 0
	for m.Modify(current) {
		i++
	}
	if i != m.GetTotalSteps() {
		t.Error(i, m.GetTotalSteps())
	}
}
