package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/helpers"
	"github.com/Erlast/short-url.git/internal/app/storages"
)

const lenString = 7

type BodyRequested struct {
	URL string `json:"url"`
}

type BodyResponse struct {
	ShortURL string `json:"result"`
}

func GetHandler(res http.ResponseWriter, req *http.Request, storage storages.URLStorage) {
	id := chi.URLParam(req, "id")

	originalURL, ok := storage.GetByID(id)

	if ok != nil {
		http.Error(res, "Not found", http.StatusNotFound)
		return
	}

	http.Redirect(res, req, originalURL, http.StatusTemporaryRedirect)
}

func PostHandler(res http.ResponseWriter, req *http.Request, storage storages.URLStorage, conf *config.Cfg) {
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

	err = storage.SaveURL(rndString, string(u))

	if err != nil {
		log.Printf("can't save url: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

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

func PostShortenHandler(res http.ResponseWriter, req *http.Request, storage storages.URLStorage, conf *config.Cfg) {
	if req.Body == http.NoBody {
		http.Error(res, "Empty String!", http.StatusBadRequest)
		return
	}

	bodyReq := BodyRequested{}

	body, err := io.ReadAll(req.Body)

	if err != nil {
		log.Printf("failed to read the request body: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &bodyReq)

	if err != nil {
		log.Printf("failed to unmarshal body: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	rndString, err := generateRandom(lenString, storage)

	if err != nil {
		log.Printf("can't generate url: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	err = storage.SaveURL(rndString, bodyReq.URL)
	if err != nil {
		log.Printf("can't save url: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	var bodyResp BodyResponse

	str, err := url.JoinPath(conf.FlagBaseURL, "/", rndString)

	if err != nil {
		http.Error(res, "Не удалось сформировать путь", http.StatusInternalServerError)
		return
	}

	bodyResp.ShortURL = str

	resp, err := json.Marshal(bodyResp)

	if err != nil {
		log.Printf("failed to marshal result: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	_, err = res.Write(resp)
	if err != nil {
		http.Error(res, "", http.StatusInternalServerError)
		return
	}
}

func generateRandom(ln int, storage storages.URLStorage) (string, error) {
	for range 3 {
		rndString := helpers.RandomString(ln)

		if !storage.IsExists(rndString) {
			return rndString, nil
		}
	}
	return "", errors.New("failed to generate a unique string")
}
