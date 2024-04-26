package main

import (
	"math/rand"
	"net/http"
	"strings"
)

func urlHandler(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodGet && req.Method != http.MethodPost {
		http.Error(res, "Method not allowed!", http.StatusMethodNotAllowed)
		return
	}

	if req.Method == http.MethodGet {
		res.WriteHeader(http.StatusTemporaryRedirect)
		res.Write([]byte("Location: https://practicum.yandex.ru/"))
	}

	if req.Method == http.MethodPost {
		str := "http://localhost:8080/" + randomString(7)
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(str))

	}

}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/test/{id}", urlHandler)
	mux.HandleFunc("/", urlHandler)

	err := http.ListenAndServe(`:8080`, mux)

	if err != nil {
		panic(err)
	}
}
