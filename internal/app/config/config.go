package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

type Cfg struct {
	flagRunAddr string
	flagBaseURL string
}

type envCfg struct {
	runAddr string `env:"SERVER_ADDRESS"`
	baseURL string `env:"BASE_URL"`
}

const defaultRunAddr = ":8080"
const defaultBaseURL = "http://localhost:8080"

func (conf *Cfg) GetBaseURL() string {
	return conf.flagBaseURL
}

func (conf *Cfg) GetRunAddr() string {
	return conf.flagRunAddr
}

func ParseFlags() *Cfg {
	config := &Cfg{
		flagRunAddr: defaultRunAddr,
		flagBaseURL: defaultBaseURL,
	}

	flag.StringVar(&config.flagRunAddr, "a", config.GetRunAddr(), "port to run server")
	flag.StringVar(&config.flagBaseURL, "b", config.GetBaseURL(), "base URL")

	flag.Parse()

	var cfg envCfg

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("can't parse")
	}

	if len(cfg.runAddr) != 0 {
		config.flagRunAddr = cfg.runAddr
	}

	if len(cfg.baseURL) != 0 {
		config.flagBaseURL = cfg.baseURL
	}

	return config
}
