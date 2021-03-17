package quote

type Repository interface {
	GetQuotes() ([]Quote, error)
	GetQuote(symbol string) (*Quote, error)
	UpdateQuoteStatus(quote *Quote, status Status) error
}
