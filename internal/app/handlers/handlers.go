package handlers

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/helpers"
	"github.com/Erlast/short-url.git/internal/app/storages"
)

const lenString = 7

func GetHandler(res http.ResponseWriter, req *http.Request, storage *storages.Storage) {
	id := chi.URLParam(req, "id")

	keyURL := strings.Split(id, "\r\n")

	originalURL, ok := storage.GetByID(keyURL[0])

	if ok != nil {
		http.Error(res, "Not found", http.StatusNotFound)
		return
	}

	http.Redirect(res, req, originalURL, http.StatusTemporaryRedirect)
}

func PostHandler(res http.ResponseWriter, req *http.Request, storage *storages.Storage, conf *config.Cfg) {
	if req.Body == http.NoBody {
		http.Error(res, "Empty String!", http.StatusBadRequest)
		return
	}

	u, err := io.ReadAll(req.Body)

	if err != nil {
		log.Printf("failed to read the request body: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	rndString, err := generateRandom(lenString, storage)

	if err != nil {
		log.Printf("can't generate url: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	storage.SaveURL(rndString, string(u))

	str, err := url.JoinPath(conf.FlagBaseURL, "/", rndString)

	if err != nil {
		http.Error(res, "Не удалось сформировать путь", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)

	_, err = res.Write([]byte(str))
	if err != nil {
		http.Error(res, "", http.StatusInternalServerError)
		return
	}
}

func generateRandom(ln int, storage *storages.Storage) (string, error) {
	for range 3 {
		rndString := helpers.RandomString(ln)

		if !storage.IsExists(rndString) {
			return rndString, nil
		}
	}
	return "", errors.New("failed to generate a unique string")
}
