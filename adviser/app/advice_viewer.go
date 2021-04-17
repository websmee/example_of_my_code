package app

import (
	"context"

	"github.com/go-echarts/go-echarts/v2/charts"

	"github.com/websmee/example_of_my_code/adviser/domain/advice"
	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
)

type ResultsViewerApp interface {
	GetCharts(ctx context.Context, advicesName string, offset, limit int) ([]*charts.Line, error)
}

type viewerApp struct {
	adviceRepository advice.Repository
	viewer           advice.Viewer
}

func NewViewerApp(
	adviceRepository advice.Repository,
	candlestickRepository candlestick.Repository,
) ResultsViewerApp {
	candlestickRepository = candlestick.NewBasicFilter(candlestickRepository)
	return &viewerApp{
		adviceRepository: adviceRepository,
		viewer:           advice.NewCBSViewer(candlestickRepository),
	}
}

func (r viewerApp) GetCharts(ctx context.Context, advicesName string, offset, limit int) ([]*charts.Line, error) {
	advices, err := r.adviceRepository.LoadAdvices(advicesName)
	if err != nil {
		return nil, err
	}

	var cs []*charts.Line
	for i := range advices[offset:limit] {
		c, err := r.viewer.GetChart(ctx, advices[i])
		if err != nil {
			return nil, err
		}

		cs = append(cs, c)
	}

	return cs, nil
}
