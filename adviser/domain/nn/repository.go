package nn

type Repository interface {
	SaveNN(name string, net *NN) error
	LoadNN(name string, net *NN) error
}
