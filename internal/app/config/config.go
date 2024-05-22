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
}

type envCfg struct {
	RunAddr     string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL     string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStorage string `env:"FILE_STORAGE_PATH" envDefault:"/tmp/short-url-db.json"`
}

const defaultRunAddr = ":8080"
const defaultBaseURL = "http://localhost:8080"
const defaultFileStoragePath = "/tmp/short-url-db.json"

func ParseFlags() *Cfg {
	config := &Cfg{}

	flag.StringVar(&config.FlagRunAddr, "a", defaultRunAddr, "port to run server")
	flag.StringVar(&config.FlagBaseURL, "b", defaultBaseURL, "base URL")
	flag.StringVar(&config.FileStorage, "f", defaultFileStoragePath, "file storage path")

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
