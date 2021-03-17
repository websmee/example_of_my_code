package quote

import "context"

type Repository interface {
	GetQuotes(ctx context.Context) ([]Quote, error)
}
