package config

import (
	"flag"
	"github.com/caarlos0/env/v11"
)

type Cfg struct {
	FlagRunAddr string
	FlagBaseURL string
}

type envCfg struct {
	runAddr string `env:"SERVER_ADDRESS"`
	baseURL string `env:"BASE_URL"`
}

func ParseFlags() (Cfg, error) {
	Config := Cfg{
		":8080",
		"http://localhost:8080",
	}

	flag.StringVar(&Config.FlagRunAddr, "a", Config.FlagRunAddr, "port to run server")
	flag.StringVar(&Config.FlagBaseURL, "b", Config.FlagBaseURL, "base URL")

	flag.Parse()

	var cfg envCfg

	err := env.Parse(&cfg)

	if len(cfg.runAddr) != 0 {
		Config.FlagRunAddr = cfg.runAddr
	}

	if len(cfg.baseURL) != 0 {
		Config.FlagBaseURL = cfg.baseURL
	}

	return Config, err

}
