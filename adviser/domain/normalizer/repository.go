package normalizer

type Repository interface {
	SaveNormalizer(name string, n *Normalizer) error
	LoadNormalizer(name string) (*Normalizer, error)
}
