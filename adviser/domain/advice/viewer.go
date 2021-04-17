package advice

import (
	"context"

	"github.com/go-echarts/go-echarts/v2/charts"
)

type Viewer interface {
	GetChart(ctx context.Context, advice InternalAdvice) (*charts.Line, error)
}
