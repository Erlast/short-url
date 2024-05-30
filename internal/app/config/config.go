package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

type Cfg struct {
	FlagRunAddr string
	FlagBaseURL string
	FileStorage string
	DatabaseDSN string
}

type envCfg struct {
	RunAddr     string `env:"SERVER_ADDRESS"`
	BaseURL     string `env:"BASE_URL"`
	FileStorage string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN string `env:"DATABASE_DSN"`
}

const defaultRunAddr = ":8080"
const defaultBaseURL = "http://localhost:8080"
const defaultFileStoragePath = "/tmp/short-url-db.json"

func ParseFlags() *Cfg {
	config := &Cfg{
		FlagRunAddr: defaultRunAddr,
		FlagBaseURL: defaultBaseURL,
		FileStorage: defaultFileStoragePath,
		DatabaseDSN: "",
	}

	flag.StringVar(&config.FlagRunAddr, "a", config.FlagRunAddr, "port to run server")
	flag.StringVar(&config.FlagBaseURL, "b", config.FlagBaseURL, "base URL")
	flag.StringVar(&config.FileStorage, "f", config.FileStorage, "file storage path")
	flag.StringVar(&config.DatabaseDSN, "d", config.DatabaseDSN, "database DSN")

	flag.Parse()
	cfg := envCfg{}

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("can't parse")
	}

	if len(cfg.RunAddr) != 0 {
		config.FlagRunAddr = cfg.RunAddr
	}

	if len(cfg.BaseURL) != 0 {
		config.FlagBaseURL = cfg.BaseURL
	}

	if len(cfg.FileStorage) != 0 {
		config.FileStorage = cfg.FileStorage
	}

	return config
}
