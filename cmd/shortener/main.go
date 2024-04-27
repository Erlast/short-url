package main

import (
	"io"
	"math/rand"
	"net/http"
	"strings"
)

var storage map[string]string

func getHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Method not allowed!", http.StatusMethodNotAllowed)
		return
	}

	id := strings.Replace(req.URL.Path, "/", "", 1)

	url, err := storage[id]

	if !err {
		http.Error(res, "Not Found!", http.StatusNotFound)
	}

	http.Redirect(res, req, string(url), http.StatusTemporaryRedirect)
}

func postHandler(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(res, "Method not allowed!", http.StatusMethodNotAllowed)
		return
	}

	defer req.Body.Close()

	if req.Body == http.NoBody {
		http.Error(res, "Empty String!", http.StatusBadRequest)
	}

	u, err := io.ReadAll(req.Body)

	if err != nil {
		http.Error(res, "Something went wrong!", http.StatusBadRequest)
	}

	rndString := randomString(7)

	storage[rndString] = string(u)

	str := "http://localhost:8080/" + rndString
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(str))

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

	mux.HandleFunc("/{id}", getHandler)
	mux.HandleFunc("/", postHandler)

	err := http.ListenAndServe(`:8080`, mux)

	if err != nil {
		panic(err)
	}
}
