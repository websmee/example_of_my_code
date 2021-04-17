package infrastructure

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/domain/advice"
	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
)

type adviceFileRepository struct {
	filePath string
}

func NewAdviceFileRepository(filePath string) advice.Repository {
	return &adviceFileRepository{
		filePath: filePath,
	}
}

func (r adviceFileRepository) SaveAdvices(name string, advices []advice.InternalAdvice) error {
	f, err := os.Create(r.getFilepath(name))
	if err != nil {
		return errors.Wrap(err, "SaveAdvices file create failed")
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	for i := range advices {
		if err := writer.Write(adviceToRecord(advices[i])); err != nil {
			return errors.Wrap(err, "SaveAdvices file write failed")
		}
	}

	return nil
}

func (r adviceFileRepository) LoadAdvices(name string) ([]advice.InternalAdvice, error) {
	f, err := os.Open(r.getFilepath(name))
	if err != nil {
		return nil, errors.Wrap(err, "LoadAdvices file open failed")
	}
	defer f.Close()

	var advices []advice.InternalAdvice
	reader := csv.NewReader(f)
	for {
		record, err := reader.Read()
		if err == io.EOF || len(record) == 0 {
			break
		}

		advices = append(advices, recordToAdvice(record))
	}

	return advices, nil
}

func adviceToRecord(advice advice.InternalAdvice) []string {
	record := []string{
		string(advice.Status),
		advice.QuoteSymbol,
		strconv.Itoa(advice.HoursBefore),
		strconv.Itoa(advice.HoursAfter),
		strconv.Itoa(int(advice.Timestamp.Unix())),
		advice.CurrentPrice.String(),
		advice.TakeProfit.String(),
		advice.StopLoss.String(),
		strconv.Itoa(advice.Leverage),
		string(advice.OrderResult),
		strconv.Itoa(int(advice.OrderClosed.Unix())),
		string(advice.AdviserType),
	}
	for i := range advice.AdviserParams {
		record = append(record, advice.AdviserParams[i].String())
	}

	return record
}

func recordToAdvice(record []string) advice.InternalAdvice {
	// todo: check errors
	status := advice.Status(record[0])
	quoteSymbol := record[1]
	hoursBefore, _ := strconv.Atoi(record[2])
	hoursAfter, _ := strconv.Atoi(record[3])
	timestamp, _ := strconv.Atoi(record[4])
	currentPrice, _ := decimal.NewFromString(record[5])
	takeProfit, _ := decimal.NewFromString(record[6])
	stopLoss, _ := decimal.NewFromString(record[7])
	leverage, _ := strconv.Atoi(record[8])
	orderResult := candlestick.OrderResult(record[9])
	orderClosed, _ := strconv.Atoi(record[10])
	adviserType := advice.AdviserType(record[11])

	adviserParams := make([]decimal.Decimal, len(record)-12)
	for i, r := range record[12:] {
		adviserParams[i], _ = decimal.NewFromString(r)
	}

	return advice.InternalAdvice{
		Status:        status,
		QuoteSymbol:   quoteSymbol,
		HoursBefore:   hoursBefore,
		HoursAfter:    hoursAfter,
		Timestamp:     time.Unix(int64(timestamp), 0),
		CurrentPrice:  currentPrice,
		TakeProfit:    takeProfit,
		StopLoss:      stopLoss,
		Leverage:      leverage,
		OrderResult:   orderResult,
		OrderClosed:   time.Unix(int64(orderClosed), 0),
		AdviserType:   adviserType,
		AdviserParams: adviserParams,
	}
}

func (r adviceFileRepository) getFilepath(name string) string {
	//todo: normalize filename
	return r.filePath + strings.ReplaceAll(name, "=", "_") + ".csv"
}
