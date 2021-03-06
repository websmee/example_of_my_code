package config

import (
	"encoding/json"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"

	configKV "github.com/websmee/ms/pkg/config"
)

const (
	tiingoKey = "tiingo"
)

type Config interface {
	GetDB(key string) (*DB, error)
	GetTiingo() (*Tiingo, error)
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
		return nil, errors.Wrap(err, "GetDB config failed get")
	}

	var db DB
	err = json.Unmarshal(data, &db)
	if err != nil {
		return nil, errors.Wrap(err, "GetDB config failed unmarshal")
	}

	return &db, nil
}

func (r *consulKVConfig) GetTiingo() (*Tiingo, error) {
	data, err := r.kv.Get(tiingoKey)
	if err != nil {
		return nil, errors.Wrap(err, "GetTiingo config failed get")
	}

	var t Tiingo
	err = json.Unmarshal(data, &t)
	if err != nil {
		return nil, errors.Wrap(err, "GetTiingo config failed unmarshal")
	}

	return &t, nil
}
