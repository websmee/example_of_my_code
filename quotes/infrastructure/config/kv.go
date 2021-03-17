package config

import (
	"encoding/json"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"

	configKV "github.com/websmee/ms/pkg/config"
)

const alphaVantageAPIKeyKey = "alpha_vantage"

type Config interface {
	GetDB(key string) (*DB, error)
	GetAlphaVantage() (*AlphaVantage, error)
}

type consulKVConfig struct {
	kv     configKV.KV
	logger log.Logger
}

func NewConsulKVConfig(consulAddr string, logger log.Logger) (Config, error) {
	kv, err := configKV.NewConsulKV(consulAddr)
	if err != nil {
		return nil, errors.Wrap(err, "NewConsulKVConfig failed")
	}

	return &consulKVConfig{kv, logger}, nil
}

func (r *consulKVConfig) GetDB(key string) (*DB, error) {
	data, err := r.kv.Get(key)
	if err != nil {
		return nil, errors.Wrap(err, "GetDB failed get")
	}

	var db DB
	err = json.Unmarshal(data, &db)
	if err != nil {
		return nil, errors.Wrap(err, "GetDB failed unmarshal")
	}

	return &db, nil
}

func (r *consulKVConfig) GetAlphaVantage() (*AlphaVantage, error) {
	data, err := r.kv.Get(alphaVantageAPIKeyKey)
	if err != nil {
		return nil, errors.Wrap(err, "GetAlphaVantageAPIKey failed get")
	}

	var av AlphaVantage
	err = json.Unmarshal(data, &av)
	if err != nil {
		return nil, errors.Wrap(err, "GetAlphaVantageAPIKey failed unmarshal")
	}

	return &av, nil
}
