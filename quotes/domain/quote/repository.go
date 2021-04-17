package quote

type Repository interface {
	GetQuotes(status Status) ([]Quote, error)
	GetQuote(symbol string) (*Quote, error)
	UpdateQuoteStatus(quote *Quote, status Status) error
}
