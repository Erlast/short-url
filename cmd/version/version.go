package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	bVersion := os.Getenv("BUILD_VERSION")
	if bVersion == "" {
		bVersion = "N/A"
	}
	bCommit := os.Getenv("BUILD_COMMIT")
	if bCommit == "" {
		bCommit = "N/A"
	}
	bDate := time.Now().Format(time.DateTime)

	code := fmt.Sprintf(`package main
var (
	BuildVersion = "%s"
	BuildDate    = "%s"
	BuildCommit  = "%s"
)
`, bVersion, bDate, bCommit)

	file, err := os.Create("v_gen.go")
	if err != nil {
		log.Printf("Error creating file: %v \n", err)
		return
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			log.Printf("Error closing file v_gen.go: %v \n", err)
		}
	}(file)

	_, err = file.WriteString(code)
	if err != nil {
		log.Printf("Error writing to file: %v \n", err)
	}
}
