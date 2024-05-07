package config

import (
	"flag"
	"log"

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

const defaultRunAddr = ":8080"
const defaultBaseURL = "http://localhost:8080"

func ParseFlags() *Cfg {
	config := &Cfg{
		FlagRunAddr: defaultRunAddr,
		FlagBaseURL: defaultBaseURL,
	}

	flag.StringVar(&config.FlagRunAddr, "a", defaultRunAddr, "port to run server")
	flag.StringVar(&config.FlagBaseURL, "b", defaultBaseURL, "base URL")

	flag.Parse()
	var cfg envCfg

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("can't parse")
	}

	if len(cfg.runAddr) != 0 {
		config.FlagRunAddr = cfg.runAddr
	}

	if len(cfg.baseURL) != 0 {
		config.FlagBaseURL = cfg.baseURL
	}

	return config
}
