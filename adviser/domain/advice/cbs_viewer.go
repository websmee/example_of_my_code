package advice

import (
	"context"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
)

type cbsViewer struct {
	candlestickRepository candlestick.Repository
}

func NewCBSViewer(candlestickRepository candlestick.Repository) Viewer {
	return &cbsViewer{candlestickRepository: candlestickRepository}
}

func (r cbsViewer) GetChart(ctx context.Context, advice InternalAdvice) (*charts.Line, error) {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: advice.QuoteSymbol + " " + string(advice.AdviserType) + " " + advice.Timestamp.Format(time.RFC3339),
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale: true,
		}),
	)

	periodBefore, err := r.candlestickRepository.GetCandlesticksByCount(
		ctx,
		advice.QuoteSymbol,
		candlestick.IntervalHour,
		advice.Timestamp,
		candlestick.GetterDirectionBackward,
		advice.HoursBefore,
	)
	if err != nil {
		return nil, err
	}

	periodAfter, err := r.candlestickRepository.GetCandlesticks(
		ctx,
		advice.QuoteSymbol,
		candlestick.IntervalHour,
		advice.Timestamp.Add(time.Hour),
		advice.OrderClosed.Add(3*time.Hour),
	)
	if err != nil {
		return nil, err
	}

	period := append(periodBefore, periodAfter...)
	hours := make([]string, len(period))
	for i := range period {
		hours[i] = period[i].Timestamp.Format("15:04")
	}

	prices := make([]opts.LineData, len(period))
	for i := range period {
		prices[i] = opts.LineData{Value: period[i].Close}
	}

	line.SetXAxis(hours).AddSeries("Hourly Close Prices", prices).
		SetSeriesOptions(
			charts.WithLabelOpts(
				opts.Label{
					Show: true,
				}),
			charts.WithAreaStyleOpts(
				opts.AreaStyle{
					Opacity: 0.2,
				}),
		)

	return line, nil
}
