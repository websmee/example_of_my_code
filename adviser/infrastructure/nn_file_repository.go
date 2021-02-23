package infrastructure

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"gonum.org/v1/gonum/mat"

	"github.com/websmee/example_of_my_code/adviser/domain/nn"
)

type NNFileRepository struct {
	filePath string
}

func NewNNFileRepository(filePath string) nn.Repository {
	return &NNFileRepository{
		filePath: filePath,
	}
}

func (r NNFileRepository) SaveNN(name string, net *nn.NN) error {
	if err := r.saveWeights(name+"-hweights", net.HiddenWeights); err != nil {
		return err
	}
	if err := r.saveWeights(name+"-oweights", net.OutputWeights); err != nil {
		return err
	}

	return nil
}

func (r NNFileRepository) saveWeights(fileName string, weights *mat.Dense) error {
	h, err := os.Create(r.getFilepath(fileName))
	if err != nil {
		return errors.Wrap(err, "saveWeights file open failed")
	}
	defer h.Close()
	if _, err := weights.MarshalBinaryTo(h); err != nil {
		return errors.Wrap(err, "saveWeights marshaling failed")
	}

	return nil
}

func (r NNFileRepository) LoadNN(name string, net *nn.NN) error {
	if err := r.loadWeights(name+"-hweights", net.HiddenWeights); err != nil {
		return err
	}
	if err := r.loadWeights(name+"-oweights", net.OutputWeights); err != nil {
		return err
	}

	return nil
}

func (r NNFileRepository) loadWeights(fileName string, weights *mat.Dense) error {
	h, err := os.Open(r.getFilepath(fileName))
	if err != nil {
		return errors.Wrap(err, "loadWeights file open failed")
	}
	defer h.Close()
	weights.Reset()
	if _, err := weights.UnmarshalBinaryFrom(h); err != nil {
		return errors.Wrap(err, "loadWeights unmarshal fail")
	}

	return nil
}

func (r NNFileRepository) getFilepath(name string) string {
	//todo: normalize filename
	return r.filePath + strings.ReplaceAll(name, "=", "_") + ".model"
}
