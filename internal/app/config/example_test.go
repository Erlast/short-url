package config

import (
	"fmt"
)

func ExampleCfg() {
	cfg := &Cfg{
		FlagRunAddr: ":8080",
		FlagBaseURL: "http://localhost:8080",
		FileStorage: "/tmp/short-url-db.json",
		SecretKey:   "secret",
		DatabaseDSN: "user:password@/dbname",
	}

	// Форматируем вывод
	output := fmt.Sprintf(
		`{
	RunAddress: %q,
	BaseURL: %q,
	FileStorage: %q,
	SecretKey: %q,
	DatabaseDSN: %q
}`,
		cfg.FlagRunAddr,
		cfg.FlagBaseURL,
		cfg.FileStorage,
		cfg.SecretKey,
		cfg.DatabaseDSN,
	)

	fmt.Println(output)
	// Output:
	// {
	//	RunAddress: ":8080",
	//	BaseURL: "http://localhost:8080",
	//	FileStorage: "/tmp/short-url-db.json",
	//	SecretKey: "secret",
	//	DatabaseDSN: "user:password@/dbname"
	// }
}
