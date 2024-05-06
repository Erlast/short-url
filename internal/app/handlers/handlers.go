package handlers

import (
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/helpers"
	"github.com/Erlast/short-url.git/internal/app/storages"
)

const lenString = 7

func GetHandler(res http.ResponseWriter, req *http.Request, storage *storages.Storage) {
	id := chi.URLParam(req, "id")

	originalURL, ok := storage.GetByID(id)

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
		http.Error(res, "Something went wrong!", http.StatusBadRequest)
		return
	}

	rndString, err := generateRandom(lenString, storage)

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	storage.SaveURL(rndString, string(u))

	str, err := url.JoinPath(conf.FlagBaseURL, "/", rndString)

	if err != nil {
		http.Error(res, "Не удалось сформировать путь", http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)

	_, err = res.Write([]byte(str))
	if err != nil {
		http.Error(res, "Something went wrong!", http.StatusInternalServerError)
		return
	}
}

func generateRandom(ln int, storage *storages.Storage) (string, error) {
	rndString := helpers.RandomString(ln)

	for range 3 {
		if _, ok := storage.GetByID(rndString); ok == nil {
			return "", errors.New("failed to generate a unique string")
		}

		rndString = helpers.RandomString(ln)
	}

	return rndString, nil
}
