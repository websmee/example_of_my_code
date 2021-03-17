package infrastructure

import (
	"bytes"
	"crypto/tls"
	"encoding/csv"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/quotes/domain/candlestick"
	"github.com/websmee/example_of_my_code/quotes/domain/quote"
	"github.com/websmee/example_of_my_code/quotes/infrastructure/config"
)

const (
	urlTemplate = "https://www.alphavantage.co/query?function=TIME_SERIES_INTRADAY_EXTENDED&symbol={symbol}&interval={interval}&slice={slice}&apikey={apikey}"
	yearsBack   = 2
)

type candlestickAlphaVantageLoader struct {
	alphaVantageConfig *config.AlphaVantage
}

func NewCandlestickAlphaVantageLoader(alphaVantageConfig *config.AlphaVantage) candlestick.Loader {
	return &candlestickAlphaVantageLoader{
		alphaVantageConfig: alphaVantageConfig,
	}
}

func (r candlestickAlphaVantageLoader) LoadCandlesticks(quote quote.Quote, start, end time.Time, interval candlestick.Interval) ([]candlestick.Candlestick, error) {
	cs := make([]candlestick.Candlestick, int(end.Sub(start).Hours())+1)
	i := 0
	for _, slice := range getAlphaVantageSlices(start, end) {
		data, err := r.getCandlesticksCSV(quote.Symbol, getAlphaVantageInterval(interval), slice)
		if err != nil {
			return nil, err
		}

		reader := csv.NewReader(bytes.NewBuffer(data))
		record, err := reader.Read() // read header
		if err == io.EOF || len(record) == 0 || record[0] == "{}" {
			return nil, errors.Wrap(err, "LoadCandlesticks csv read failed slice "+slice)
		}

		for {
			record, err := reader.Read()     // read next line
			if err != nil && err != io.EOF { // some read error
				return nil, errors.Wrap(err, "LoadCandlesticks csv read failed slice "+slice)
			}
			if err == io.EOF || len(record) == 0 { // end of the range
				break
			}
			if record[0] == "" { // skip empty
				continue
			}

			c, err := alphaVantageCSVRecordToCandlestick(record, quote, interval)
			if err != nil {
				return nil, errors.WithMessage(err, "LoadCandlesticks failed parsing slice "+slice)
			}

			cs[i] = c
			i++
		}
		time.Sleep(20 * time.Second) // free account restriction
	}

	return cs[:i], nil
}

func (r candlestickAlphaVantageLoader) getCandlesticksCSV(symbol, interval, slice string) ([]byte, error) {
	url := getAlphaVantageURL(symbol, interval, slice, r.alphaVantageConfig.APIKey)
	req, _ := http.NewRequest(http.MethodGet, url, nil)

	http.DefaultClient.Transport = &http.Transport{
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "getCandlesticksCSV cannot download data")
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return body, nil
}

func getAlphaVantageURL(symbol, interval, slice, apikey string) string {
	url := strings.Replace(urlTemplate, "{apikey}", apikey, 1)
	url = strings.Replace(url, "{symbol}", symbol, 1)
	url = strings.Replace(url, "{interval}", interval, 1)
	url = strings.Replace(url, "{slice}", slice, 1)

	return url
}

func getAlphaVantageInterval(interval candlestick.Interval) string {
	if interval == candlestick.IntervalHour {
		return "60min"
	}

	return ""
}

func getAlphaVantageSlices(_, _ time.Time) []string {
	var slices []string
	for y := 1; y <= yearsBack; y++ {
		for m := 1; m <= 12; m++ {
			slices = append(slices, "year"+strconv.Itoa(y)+"month"+strconv.Itoa(m))
		}
	}

	return slices
}

func alphaVantageCSVRecordToCandlestick(record []string, quote quote.Quote, interval candlestick.Interval) (c candlestick.Candlestick, err error) {
	c.Interval = interval
	c.QuoteID = quote.ID
	c.AdjClose = decimal.NewFromInt(0)

	c.Timestamp, err = time.Parse("2006-01-02 15:04:05", record[0])
	if err != nil {
		return c, errors.Wrap(err, "alphaVantageCSVRecordToCandlestick cannot convert date")
	}

	c.Open, err = decimal.NewFromString(record[1])
	if err != nil {
		return c, errors.Wrap(err, "alphaVantageCSVRecordToCandlestick cannot convert open price")
	}

	c.High, err = decimal.NewFromString(record[2])
	if err != nil {
		return c, errors.Wrap(err, "alphaVantageCSVRecordToCandlestick cannot convert high price")
	}

	c.Low, err = decimal.NewFromString(record[3])
	if err != nil {
		return c, errors.Wrap(err, "alphaVantageCSVRecordToCandlestick cannot convert low price")
	}

	c.Close, err = decimal.NewFromString(record[4])
	if err != nil {
		return c, errors.Wrap(err, "alphaVantageCSVRecordToCandlestick cannot convert close price")
	}

	c.Volume, err = strconv.Atoi(record[5])
	if err != nil {
		return c, errors.Wrap(err, "alphaVantageCSVRecordToCandlestick cannot convert volume")
	}

	return c, nil
}
