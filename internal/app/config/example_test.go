package config

import (
	"fmt"
)

func ExampleConfig() {
	config := &Cfg{
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
		config.FlagRunAddr,
		config.FlagBaseURL,
		config.FileStorage,
		config.SecretKey,
		config.DatabaseDSN,
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
