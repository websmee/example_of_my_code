package advice

type Repository interface {
	SaveAdvices(name string, advices []InternalAdvice) error
	LoadAdvices(name string) ([]InternalAdvice, error)
}
