package config

import (
	"flag"
)

var FlagRunAddr string
var FlagBaseUrl string

func ParseFlags() {

	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&FlagBaseUrl, "b", "localhost:8080", "base URL")

	flag.Parse()

}
