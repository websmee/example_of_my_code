package app

import (
	"context"
	"math/rand"
	"time"

	"github.com/websmee/example_of_my_code/adviser/domain/dataset"
	"github.com/websmee/example_of_my_code/adviser/domain/nn"
)

type NNTrainerApp interface {
	TrainNN(ctx context.Context, nnName, datasetName string, epochs int, learningRate float64) error
}

type nnTrainerApp struct {
	nnRepository      nn.Repository
	datasetRepository dataset.Repository
	trainer           nn.Trainer
}

func NewNNTrainerApp(
	nnRepository nn.Repository,
	datasetRepository dataset.Repository,
) NNTrainerApp {
	return &nnTrainerApp{
		nnRepository:      nnRepository,
		datasetRepository: datasetRepository,
		trainer:           nn.DefaultTrainer(nn.DefaultCalculator()),
	}
}

func (r nnTrainerApp) TrainNN(_ context.Context, nnName, datasetName string, epochs int, learningRate float64) error {
	dsTrain := dataset.NewCBSDataset()
	if err := r.datasetRepository.LoadDataset(datasetName, dsTrain); err != nil {
		return err
	}

	net := nn.NewCBSNN()
	r.train(net, dsTrain, epochs, learningRate)

	return r.nnRepository.SaveNN(nnName, net)
}

func (r nnTrainerApp) train(net *nn.NN, ds *dataset.Dataset, epochs int, learningRate float64) {
	rand.Seed(time.Now().UTC().UnixNano())
	for e := 0; e < epochs; e++ {
		for i := range ds.Rows {
			inputs := make([]float64, net.Inputs)
			for j := range inputs {
				x, _ := ds.Rows[i].NormalizedInputs[j].Float64()
				inputs[j] = x
			}

			outputs := make([]float64, net.Outputs)
			for j := range outputs {
				x, _ := ds.Rows[i].Outputs[j].Float64()
				outputs[j] = x
			}

			r.trainer.Train(net, inputs, outputs, learningRate)
		}
	}
}
