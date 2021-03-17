package infrastructure

import (
	"context"

	"github.com/websmee/example_of_my_code/adviser/domain/quote"
	"github.com/websmee/example_of_my_code/adviser/infrastructure/grpc"
)

type quoteGRPCRepository struct {
	quotesApp grpc.QuotesApp
}

func NewQuoteGRPCRepository(
	quotesApp grpc.QuotesApp,
) quote.Repository {
	return &quoteGRPCRepository{
		quotesApp: quotesApp,
	}
}

func (r quoteGRPCRepository) GetQuotes(ctx context.Context) ([]quote.Quote, error) {
	return r.quotesApp.GetQuotes(ctx)
}
