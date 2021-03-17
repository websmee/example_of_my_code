package persistence

import (
	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"

	"github.com/websmee/example_of_my_code/quotes/domain/quote"
)

type QuoteRepository struct {
	db *pg.DB
}

func NewQuoteRepository(db *pg.DB) *QuoteRepository {
	return &QuoteRepository{db}
}

func (r QuoteRepository) GetQuotes() ([]quote.Quote, error) {
	var quotes []quote.Quote

	err := r.db.Model(&quote.Quote{}).Select(&quotes)

	if err != nil && err != pg.ErrNoRows {
		return nil, errors.Wrap(err, "GetQuotes failed")
	}

	return quotes, nil
}

func (r QuoteRepository) GetQuote(symbol string) (*quote.Quote, error) {
	q := &quote.Quote{}
	err := r.db.Model(q).Where("symbol = ?", symbol).Select()

	if err != nil && err != pg.ErrNoRows {
		return nil, errors.Wrap(err, "GetQuote failed")
	}

	return q, nil
}

func (r QuoteRepository) UpdateQuoteStatus(quote *quote.Quote, status quote.Status) error {
	_, err := r.db.Model(quote).
		Set("status = ?", status).
		WherePK().
		Update()

	return errors.Wrap(err, "UpdateQuoteStatus failed")
}
