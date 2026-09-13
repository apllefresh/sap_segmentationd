package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ConnURI          string `envconfig:"CONN_URI" default:"http://bsm.api.iql.ru/ords/bsm/segmentation/get_segmentation"`
	ConnAuthLoginPwd string `envconfig:"CONN_AUTH_LOGIN_PWD" default:"4Dfddf5:jKlljHGH"`
	ConnUserAgent    string `envconfig:"CONN_USER_AGENT" default:"spacecount-test"`
	ConnTimeout      int    `envconfig:"CONN_TIMEOUT" default:"5"`
	ImportBatchSize  int    `envconfig:"IMPORT_BATCH_SIZE" default:"50"`
}

func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
