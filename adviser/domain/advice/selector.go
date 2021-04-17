package advice

type Selector interface {
	SelectAdvice(advices []InternalAdvice) *InternalAdvice
}

type defaultSelector struct{}

func NewDefaultSelector() Selector {
	return &defaultSelector{}
}

func (r defaultSelector) SelectAdvice(advices []InternalAdvice) *InternalAdvice {
	if len(advices) == 0 {
		return nil
	}

	minPeriod := advices[0].HoursBefore
	var minPeriodAdvices []InternalAdvice
	for i := range advices {
		if minPeriod > advices[i].HoursBefore {
			minPeriod = advices[i].HoursBefore
		}
		if advices[i].HoursBefore == minPeriod {
			minPeriodAdvices = append(minPeriodAdvices, advices[i])
		}
	}

	if len(minPeriodAdvices) == 0 {
		return nil
	}

	minTakeProfitAdvice := minPeriodAdvices[0]
	for i := range minPeriodAdvices {
		if minTakeProfitAdvice.TakeProfit.GreaterThan(minPeriodAdvices[i].TakeProfit) {
			minTakeProfitAdvice = minPeriodAdvices[i]
		}
	}

	return &minTakeProfitAdvice
}
