package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name                string
		args                []string
		envVars             map[string]string
		expectedAddr        string
		expectedBaseURL     string
		expectedFileStorage string
		expectedDatabaseDSN string
		expectedSecretKey   string
	}{
		{
			name: "Default values",
			args: []string{},
			envVars: map[string]string{
				"SERVER_ADDRESS": ":8081",
			},
			expectedAddr:        ":8081",
			expectedBaseURL:     defaultBaseURL,
			expectedFileStorage: defaultFileStoragePath,
			expectedDatabaseDSN: "",
			expectedSecretKey:   "",
		},
		{
			name: "Flag values",
			args: []string{
				"-a",
				"127.0.0.1:9000",
				"-b",
				"http://localhost:9000",
				"-f",
				"/tmp/short-url-db-example.json",
				"-d",
				"postgres://user:password@localhost:5432/db",
				"-k",
				"anothersuperkey",
			},
			envVars:             map[string]string{},
			expectedAddr:        "127.0.0.1:9000",
			expectedBaseURL:     "http://localhost:9000",
			expectedFileStorage: "/tmp/short-url-db-example.json",
			expectedDatabaseDSN: "postgres://user:password@localhost:5432/db",
			expectedSecretKey:   "anothersuperkey",
		},
		{
			name: "Environment values",
			args: []string{},
			envVars: map[string]string{
				"SERVER_ADDRESS":    ":8081",
				"BASE_URL":          "http://localhost:8081",
				"FILE_STORAGE_PATH": "/tmp/short-url-db-1.json",
				"DATABASE_DSN":      "postgres://user:password@localhost:5432/db",
				"SECRET_KEY":        "supersecretkey2",
			},
			expectedAddr:        ":8081",
			expectedBaseURL:     "http://localhost:8081",
			expectedFileStorage: "/tmp/short-url-db-1.json",
			expectedDatabaseDSN: "",
			expectedSecretKey:   "supersecretkey2",
		},
		{
			name: "Flag overrides environment",
			args: []string{"-a", "127.0.0.1:9000"},
			envVars: map[string]string{
				"SERVER_ADDRESS":    ":8081",
				"BASE_URL":          "http://localhost:8081",
				"FILE_STORAGE_PATH": "/tmp/short-url-db-1.json",
				"DATABASE_DSN":      "postgres://user:password@localhost:5432/db",
				"SECRET_KEY":        "supersecretkey2",
			},
			expectedAddr:        ":8081",
			expectedBaseURL:     "http://localhost:8081",
			expectedFileStorage: "/tmp/short-url-db-1.json",
			expectedDatabaseDSN: "",
			expectedSecretKey:   "supersecretkey2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) //nolint:reassign //ось такая ось
			var cleanup []func()
			for k, v := range tt.envVars {
				if err := os.Setenv(k, v); err != nil {
					t.Fatalf("failed to set env var: %v", err)
				}
				cleanup = append(cleanup, func(key string) func() {
					return func() {
						if err := os.Unsetenv(key); err != nil {
							t.Fatalf("failed to unset env var: %v", err)
						}
					}
				}(k))
			}

			defer func() {
				for _, fn := range cleanup {
					fn()
				}
			}()

			os.Args = append([]string{os.Args[0]}, tt.args...) //nolint:reassign //ось такая ось

			config := ParseFlags()

			assert.Equal(t, tt.expectedAddr, config.FlagRunAddr)
			assert.Equal(t, tt.expectedBaseURL, config.FlagBaseURL)
			assert.Equal(t, tt.expectedFileStorage, config.FileStorage)
			assert.Equal(t, tt.expectedDatabaseDSN, config.DatabaseDSN)
			assert.Equal(t, tt.expectedSecretKey, config.SecretKey)
		})
	}
}
