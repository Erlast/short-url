package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
)

func urlHandler(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodGet && req.Method != http.MethodPost {
		http.Error(res, "Method not allowed!", http.StatusMethodNotAllowed)
		return
	}

	if req.Method == http.MethodGet {
		param := req.PathValue("id")
		fmt.Printf("Greeting received: %v\n", param)
		fmt.Printf("Greeting received: %v\n", req.URL.String())
		res.WriteHeader(http.StatusTemporaryRedirect)
		res.Write([]byte("Location: https://practicum.yandex.ru/"))
	}

	if req.Method == http.MethodPost {
		headerContentTtype := req.Header.Get("Content-Type")

		if headerContentTtype != "text/plain" {
			http.Error(res, "Content-Type is not text/plain", http.StatusUnsupportedMediaType)
			return
		}

		if req.ContentLength == 0 {
			http.Error(res, "No URL posted", http.StatusBadRequest)
			return
		}

		_, err := io.ReadAll(req.Body)

		if err != nil {
			http.Error(res, "Error parse body", http.StatusBadRequest)
			return
		}

		str := "http://localhost:8080/" + randomString(7)
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

	mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/test/123", nil))

	err := http.ListenAndServe(`:8080`, mux)

	if err != nil {
		panic(err)
	}
}
