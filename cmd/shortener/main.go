package main

import (
	"io"
	"math/rand"
	"net/http"
	"strings"
)

var storage map[string]string

func urlHandler(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodGet && req.Method != http.MethodPost {
		http.Error(res, "Method not allowed!", http.StatusMethodNotAllowed)
		return
	}

	if req.Method == http.MethodGet {
		id := strings.Replace(req.URL.Path, "/", "", 1)

		http.Redirect(res, req, string(storage[id]), http.StatusTemporaryRedirect)
	}

	if req.Method == http.MethodPost {
		u, err := io.ReadAll(req.Body)

		if err != nil {
			http.Error(res, "Empty String!", http.StatusBadRequest)
		}

		rndString := randomString(7)

		storage[rndString] = string(u)

		str := "http://localhost:8080/" + rndString
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
	storage = make(map[string]string)

	mux := http.NewServeMux()

	mux.HandleFunc("/test/{id}", urlHandler)
	mux.HandleFunc("/", urlHandler)

	err := http.ListenAndServe(`:8080`, mux)

	if err != nil {
		panic(err)
	}
}
