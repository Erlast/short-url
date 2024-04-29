package main

import (
	"github.com/go-chi/chi/v5"
	"io"
	"math/rand"
	"net/http"
	"strings"
)

var storage map[string]string

func getHandler(res http.ResponseWriter, req *http.Request) {
	id := strings.Replace(req.URL.Path, "/", "", 1)

	url, err := storage[id]

	if !err {
		http.Error(res, "Not Found!", http.StatusNotFound)
	}

	http.Redirect(res, req, url, http.StatusTemporaryRedirect)
}

func postHandler(res http.ResponseWriter, req *http.Request) {

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			http.Error(res, "Enable to close body!", http.StatusBadRequest)
		}
	}(req.Body)

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
	_, err = res.Write([]byte(str))
	if err != nil {
		return
	}

}

func checkHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet && req.Method != http.MethodPost {
		http.Error(res, "Method not allowed!", http.StatusMethodNotAllowed)
		return
	}

	if req.Method == http.MethodPost {
		postHandler(res, req)
	}

	if req.Method == http.MethodGet {
		getHandler(res, req)
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

	r := chi.NewRouter()
	r.Get("/{id}", checkHandler)

	r.Post("/", checkHandler)

	err := http.ListenAndServe(`:8080`, r)

	if err != nil {
		panic(err)
	}
}
