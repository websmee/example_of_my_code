package config

import (
	"encoding/json"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"

	configKV "github.com/websmee/ms/pkg/config"
)

type Config interface {
	GetDb(key string) (*Db, error)
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

func (r *consulKVConfig) GetDb(key string) (*Db, error) {
	data, err := r.kv.Get(key)
	if err != nil {
		return nil, errors.Wrap(err, "GetDb failed get")
	}

	var db Db
	err = json.Unmarshal(data, &db)
	if err != nil {
		return nil, errors.Wrap(err, "GetDb failed unmarshal")
	}

	return &db, nil
}
