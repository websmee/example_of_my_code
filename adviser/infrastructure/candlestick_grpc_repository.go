package infrastructure

import (
	"context"
	"time"

	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
	"github.com/websmee/example_of_my_code/adviser/infrastructure/grpc"
)

type candlestickGRPCRepository struct {
	quotesApp grpc.QuotesApp
}

func NewCandlestickGRPCRepository(
	quotesApp grpc.QuotesApp,
) candlestick.Repository {
	return &candlestickGRPCRepository{
		quotesApp: quotesApp,
	}
}

func (r candlestickGRPCRepository) GetCandlesticks(ctx context.Context, symbol string, interval candlestick.Interval, from, to time.Time) ([]candlestick.Candlestick, error) {
	return r.quotesApp.GetCandlesticks(ctx, symbol, interval, from, to)
}
