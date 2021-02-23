package dataset

type Repository interface {
	SaveDataset(name string, dataset *Dataset) error
	LoadDataset(name string, dataset *Dataset) error
}
