package config

import (
	"flag"
)

type Cfg struct {
	FlagRunAddr string
	FlagBaseUrl string
}

func ParseFlags() Cfg {
	Config := Cfg{
		":8080",
		"http://localhost:8080",
	}

	flag.StringVar(&Config.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&Config.FlagBaseUrl, "b", "localhost:8080", "base URL")

	flag.Parse()

	return Config

}
