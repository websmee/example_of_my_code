package tiingo

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	"github.com/websmee/example_of_my_code/quotes/infrastructure/config"
)

type Client interface {
	GetPrices(request PricesRequest) ([]Prices, error)
}

const baseURL = "https://api.tiingo.com"

type client struct {
	httpClient http.Client
	token      string
}

func NewClient(cfg *config.Tiingo) Client {
	return &client{
		httpClient: http.Client{},
		token:      cfg.Token,
	}
}

func (r client) GetPrices(request PricesRequest) ([]Prices, error) {
	responseBody, err := r.makeRequest(http.MethodGet, request.GetPath(), nil)
	if err != nil {
		return nil, err
	}

	var response []Prices
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, errors.Wrap(err, "GetPrices unmarshal failed")
	}

	return response, nil
}

func (r client) makeRequest(method, path string, requestBody io.Reader) ([]byte, error) {
	url := baseURL + path + "&token=" + r.token
	request, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return nil, errors.Wrap(err, "makeRequest failed")
	}

	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/json")

	response, err := r.httpClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "makeRequest do failed")
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "makeRequest read failed")
	}

	return responseBody, nil
}
