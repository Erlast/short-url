package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

// Cfg структура конфигурации.
type Cfg struct {
	FlagRunAddr string
	FlagBaseURL string
	FileStorage string
	DatabaseDSN string
	SecretKey   string
}

type envCfg struct {
	RunAddr     string `env:"SERVER_ADDRESS"`
	BaseURL     string `env:"BASE_URL"`
	FileStorage string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN string `env:"DATABASE_DSN"`
	SecretKey   string `env:"SECRET_KEY"`
}

const defaultRunAddr = ":8080"                          // defaultRunAddr порт по умолчанию
const defaultBaseURL = "http://localhost:8080"          // defaultBaseURL базовый URL приложения
const defaultFileStoragePath = "/tmp/short-url-db.json" // defaultFileStoragePath файл хранилище
const secretKey = "supersecretkey"                      // secretKey  секретный ключ для формирования jwt токенов

// ParseFlags функция разбора заданных параметров приложения.
func ParseFlags() *Cfg {
	config := &Cfg{
		FlagRunAddr: defaultRunAddr,
		FlagBaseURL: defaultBaseURL,
		FileStorage: defaultFileStoragePath,
		DatabaseDSN: "",
		SecretKey:   secretKey,
	}

	flag.StringVar(&config.FlagRunAddr, "a", config.FlagRunAddr, "port to run server")
	flag.StringVar(&config.FlagBaseURL, "b", config.FlagBaseURL, "base URL")
	flag.StringVar(&config.FileStorage, "f", config.FileStorage, "file storage path")
	flag.StringVar(&config.DatabaseDSN, "d", config.DatabaseDSN, "database DSN")
	flag.StringVar(&config.SecretKey, "k", config.DatabaseDSN, "secret key")

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

	if len(cfg.SecretKey) != 0 {
		config.SecretKey = cfg.SecretKey
	}

	return config
}
