package app

import (
	"context"
	"fmt"
	"time"

	"github.com/websmee/example_of_my_code/adviser/domain/dataset"
	"github.com/websmee/example_of_my_code/adviser/domain/nn"
)

type NNTesterApp interface {
	TestNN(ctx context.Context, nnName, datasetName string) error
}

type nnTesterApp struct {
	nnRepository      nn.Repository
	datasetRepository dataset.Repository
	predictor         nn.Predictor
}

func NewNNTesterApp(
	nnRepository nn.Repository,
	datasetRepository dataset.Repository,
) NNTesterApp {
	return &nnTesterApp{
		nnRepository:      nnRepository,
		datasetRepository: datasetRepository,
		predictor:         nn.DefaultPredictor(nn.DefaultCalculator()),
	}
}

func (r nnTesterApp) TestNN(_ context.Context, nnName, datasetName string) error {
	dsTest := dataset.NewCBSDataset()
	if err := r.datasetRepository.LoadDataset(datasetName, dsTest); err != nil {
		return err
	}

	net := nn.NewCBSNN()
	if err := r.nnRepository.LoadNN(nnName, net); err != nil {
		return err
	}

	r.test(net, dsTest)

	return nil
}

func (r nnTesterApp) test(net *nn.NN, ds *dataset.Dataset) {
	profitBuysPredicted := 0
	profitBuysPredictedConfidence90 := 0
	profitBuysPredictedConfidence70 := 0
	profitBuysPredictedConfidence50 := 0

	profitBuysScore := 0
	profitBuysScoreConfidence90 := 0
	profitBuysScoreConfidence70 := 0
	profitBuysScoreConfidence50 := 0

	profitSellsPredicted := 0
	profitSellsPredictedConfidence90 := 0
	profitSellsPredictedConfidence70 := 0
	profitSellsPredictedConfidence50 := 0

	profitSellsScore := 0
	profitSellsScoreConfidence90 := 0
	profitSellsScoreConfidence70 := 0
	profitSellsScoreConfidence50 := 0

	stopLossesPredicted := 0
	stopLossesPredictedConfidence90 := 0
	stopLossesPredictedConfidence70 := 0
	stopLossesPredictedConfidence50 := 0

	stopLossesScore := 0
	stopLossesScoreConfidence90 := 0
	stopLossesScoreConfidence70 := 0
	stopLossesScoreConfidence50 := 0

	for i := range ds.Rows {
		inputs := make([]float64, net.Inputs)
		for j := range inputs {
			x, _ := ds.Rows[i].NormalizedInputs[j].Float64()
			inputs[j] = x
		}

		predictedOutputs := r.predictor.Predict(net, inputs)
		actualOutputs := make([]float64, net.Outputs)
		for j := range actualOutputs {
			x, _ := ds.Rows[i].Outputs[j].Float64()
			actualOutputs[j] = x
		}

		predictedBest := 0
		predictedHighest := 0.0
		for j := 0; j < net.Outputs; j++ {
			if predictedOutputs.At(j, 0) > predictedHighest {
				predictedBest = j
				predictedHighest = predictedOutputs.At(j, 0)
			}
		}

		actualBest := -1
		actualHighest := 0.0
		for j := 0; j < net.Outputs; j++ {
			if actualOutputs[j] != 0.01 && actualOutputs[j] > actualHighest {
				actualBest = j
				actualHighest = actualOutputs[j]
			}
		}
		switch predictedBest {
		case 0:
			profitBuysPredicted++
			if predictedHighest >= 0.9 {
				profitBuysPredictedConfidence90++
			}
			if predictedHighest >= 0.7 {
				profitBuysPredictedConfidence70++
			}
			if predictedHighest >= 0.5 {
				profitBuysPredictedConfidence50++
			}
		case 1:
			profitSellsPredicted++
			if predictedHighest >= 0.9 {
				profitSellsPredictedConfidence90++
			}
			if predictedHighest >= 0.7 {
				profitSellsPredictedConfidence70++
			}
			if predictedHighest >= 0.5 {
				profitSellsPredictedConfidence50++
			}
		case 2:
			stopLossesPredicted++
			if predictedHighest >= 0.9 {
				stopLossesPredictedConfidence90++
			}
			if predictedHighest >= 0.7 {
				stopLossesPredictedConfidence70++
			}
			if predictedHighest >= 0.5 {
				stopLossesPredictedConfidence50++
			}
		}
		if predictedBest == actualBest {
			switch actualBest {
			case 0:
				profitBuysScore++
				if predictedHighest >= 0.9 {
					profitBuysScoreConfidence90++
					fmt.Println(
						"BUY at",
						time.Unix(ds.Rows[i].Debug[0].IntPart(), 0).UTC().Add(-5*time.Hour).Format(time.RFC3339),
						"price:",
						ds.Rows[i].Debug[1].String(),
						"tp:",
						ds.Rows[i].Debug[2].String(),
						"sl:",
						ds.Rows[i].Debug[3].String(),
					)
				}
				if predictedHighest >= 0.7 {
					profitBuysScoreConfidence70++
					fmt.Println(
						"BUY at",
						time.Unix(ds.Rows[i].Debug[0].IntPart(), 0).UTC().Add(-5*time.Hour).Format(time.RFC3339),
						"price:",
						ds.Rows[i].Debug[1].String(),
						"tp:",
						ds.Rows[i].Debug[2].String(),
						"sl:",
						ds.Rows[i].Debug[3].String(),
					)
				}
				if predictedHighest >= 0.5 {
					profitBuysScoreConfidence50++
				}
			case 1:
				profitSellsScore++
				if predictedHighest >= 0.9 {
					profitSellsScoreConfidence90++
				}
				if predictedHighest >= 0.7 {
					profitSellsScoreConfidence70++
				}
				if predictedHighest >= 0.5 {
					profitSellsScoreConfidence50++
				}
			case 2:
				stopLossesScore++
				if predictedHighest >= 0.9 {
					stopLossesScoreConfidence90++
				}
				if predictedHighest >= 0.7 {
					stopLossesScoreConfidence70++
				}
				if predictedHighest >= 0.5 {
					stopLossesScoreConfidence50++
				}
			}
		}
	}

	fmt.Println("[profit buys total]")
	verbose(float64(profitBuysPredicted), float64(profitBuysScore))
	fmt.Println("[profit buys confidence 90%]")
	verbose(float64(profitBuysPredictedConfidence90), float64(profitBuysScoreConfidence90))
	fmt.Println("[profit buys confidence 70%]")
	verbose(float64(profitBuysPredictedConfidence70), float64(profitBuysScoreConfidence70))
	fmt.Println("[profit buys confidence 50%]")
	verbose(float64(profitBuysPredictedConfidence50), float64(profitBuysScoreConfidence50))

	fmt.Println("[profit sells total]")
	verbose(float64(profitSellsPredicted), float64(profitSellsScore))
	fmt.Println("[profit sells confidence 90%]")
	verbose(float64(profitSellsPredictedConfidence90), float64(profitSellsScoreConfidence90))
	fmt.Println("[profit sells confidence 70%]")
	verbose(float64(profitSellsPredictedConfidence70), float64(profitSellsScoreConfidence70))
	fmt.Println("[profit sells confidence 50%]")
	verbose(float64(profitSellsPredictedConfidence50), float64(profitSellsScoreConfidence50))

	fmt.Println("[stop losses total]")
	verbose(float64(stopLossesPredicted), float64(stopLossesScore))
	fmt.Println("[stop losses confidence 90%]")
	verbose(float64(stopLossesPredictedConfidence90), float64(stopLossesScoreConfidence90))
	fmt.Println("[stop losses confidence 70%]")
	verbose(float64(stopLossesPredictedConfidence70), float64(stopLossesScoreConfidence70))
	fmt.Println("[stop losses confidence 50%]")
	verbose(float64(stopLossesPredictedConfidence50), float64(stopLossesScoreConfidence50))
}

func verbose(predicted, score float64) {
	effectiveness := 0.0
	if predicted != 0 {
		effectiveness = score / predicted * 100
	}
	fmt.Println("predicted:", predicted)
	fmt.Println("score:", score)
	fmt.Println("effectiveness:", effectiveness)
	fmt.Println()
}
