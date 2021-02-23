package quote

type Repository interface {
	GetQuotes() ([]Quote, error)
	GetQuote(symbol string) (*Quote, error)
}
