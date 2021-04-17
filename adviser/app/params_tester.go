package app

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"

	"github.com/websmee/example_of_my_code/adviser/domain/advice"
	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
	"github.com/websmee/example_of_my_code/adviser/domain/params"
	"github.com/websmee/example_of_my_code/adviser/domain/quote"
)

type ParamsTesterApp interface {
	TestParams(ctx context.Context, name string, from, to time.Time) error
}

type testerApp struct {
	tester           params.AdviserParamsTester
	adviser          advice.Adviser
	quoteRepository  quote.Repository
	paramsRepository params.Repository
	adviceRepository advice.Repository
	adviceSelector   advice.Selector
}

func NewCBSScaledTesterApp(
	quoteRepository quote.Repository,
	candlestickRepository candlestick.Repository,
	paramsRepository params.Repository,
	adviceRepository advice.Repository,
) ParamsTesterApp {
	return newTesterApp(
		quoteRepository,
		candlestickRepository,
		paramsRepository,
		adviceRepository,
		advice.NewCBSScaledAdviser(candlestickRepository, candlestick.NewDefaultCalculator()),
	)
}

func newTesterApp(
	quoteRepository quote.Repository,
	candlestickRepository candlestick.Repository,
	paramsRepository params.Repository,
	adviceRepository advice.Repository,
	adviser advice.Adviser,
) ParamsTesterApp {
	candlestickRepository = candlestick.NewBasicFilter(candlestickRepository)
	return &testerApp{
		tester:           params.NewAdviserParamsTester(candlestickRepository),
		adviser:          adviser,
		quoteRepository:  quoteRepository,
		paramsRepository: paramsRepository,
		adviceRepository: adviceRepository,
		adviceSelector:   advice.NewDefaultSelector(),
	}
}

func (r testerApp) TestParams(ctx context.Context, name string, from, to time.Time) error {
	p, err := r.paramsRepository.LoadParams(name)
	if err != nil {
		return err
	}

	quotes, err := r.quoteRepository.GetQuotes(ctx)
	if err != nil {
		return err
	}

	total, err := r.getTotalSteps(ctx, quotes, from, to)
	if err != nil {
		return err
	}
	bar := pb.StartNew(total)

	var count, advicesOK, profitBuy, profitSell, lossBuy, lossSell int
	var profitAdvices, lossAdvices, expiredAdvices []advice.InternalAdvice
	var wg sync.WaitGroup
	statuses := make(map[advice.Status]int64)
	advicesChan := make(chan []advice.InternalAdvice)
	for i := range quotes {
		q := quotes[i]
		wg.Add(1)
		go func() {
			r.tester.TestParams(ctx, r.adviser, p, q, from, to, advicesChan)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(advicesChan)
		bar.SetCurrent(int64(count))
	}()

	for a := range advicesChan {
		count++
		if count%(total/100) == 0 {
			bar.SetCurrent(int64(count))
		}

		var okAdvices []advice.InternalAdvice
		for i := range a {
			statuses[a[i].Status]++
			if a[i].Status == advice.StatusOK {
				okAdvices = append(okAdvices, a[i])
			}
		}

		if selectedAdvice := r.adviceSelector.SelectAdvice(okAdvices); selectedAdvice != nil {
			advicesOK++
			switch selectedAdvice.OrderResult {
			case candlestick.OrderResultTakeProfit:
				profitAdvices = append(profitAdvices, *selectedAdvice)
				if selectedAdvice.TakeProfit.GreaterThan(selectedAdvice.CurrentPrice) {
					profitBuy++
				} else {
					profitSell++
				}
			case candlestick.OrderResultStopLoss:
				lossAdvices = append(lossAdvices, *selectedAdvice)
				if selectedAdvice.TakeProfit.GreaterThan(selectedAdvice.CurrentPrice) {
					lossBuy++
				} else {
					lossSell++
				}
			case candlestick.OrderResultExpired:
				expiredAdvices = append(expiredAdvices, *selectedAdvice)
			}
		}
	}
	bar.Finish()

	frequency := strconv.FormatFloat(float64(advicesOK)/float64(count)*100, 'f', 2, 64)
	accuracy := strconv.FormatFloat(float64(len(profitAdvices))/float64(advicesOK)*100, 'f', 2, 64)
	r.printStatuses(statuses)
	fmt.Println("TOTAL", total)
	fmt.Println("PROFIT BUY", profitBuy)
	fmt.Println("PROFIT SELL", profitSell)
	fmt.Println("LOSS BUY", lossBuy)
	fmt.Println("LOSS SELL", lossSell)
	fmt.Println("EXPIRED", len(expiredAdvices))
	fmt.Println("FREQUENCY", frequency)
	fmt.Println("ACCURACY", accuracy)

	err = r.adviceRepository.SaveAdvices(name+"_profit", profitAdvices)
	if err != nil {
		return err
	}

	err = r.adviceRepository.SaveAdvices(name+"_loss", lossAdvices)
	if err != nil {
		return err
	}

	return nil
}

func (r testerApp) getTotalSteps(ctx context.Context, quotes []quote.Quote, from, to time.Time) (int, error) {
	total := 0
	for i := range quotes {
		t, err := r.tester.GetTotalSteps(ctx, quotes[i], from, to)
		if err != nil {
			return 0, err
		}
		total += t
	}

	return total, nil
}

func (r testerApp) printStatuses(statuses map[advice.Status]int64) {
	s := make([]string, len(statuses))
	i := 0
	for k := range statuses {
		s[i] = string(k)
		i++
	}
	sort.Strings(s)

	for _, status := range s {
		fmt.Println(status, statuses[advice.Status(status)])
	}
}
