package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

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

type Pinger interface {
	CheckPing() error
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

	res.Header().Set("Content-Type", "text/plain")

	rndURL, err := generateURLAndSave(lenString, storage, string(u))

	if errors.Is(err, helpers.ErrConflict) {
		res.WriteHeader(http.StatusConflict)
		str, err := url.JoinPath(conf.FlagBaseURL, "/", rndURL)

		if err != nil {
			log.Printf("can't join path %v", err)
			http.Error(res, "", http.StatusInternalServerError)
			return
		}
		_, err = res.Write([]byte(str))

		if err != nil {
			log.Printf("can't generate short url for original url %v", err)
			http.Error(res, "", http.StatusBadRequest)
			return
		}
		return
	}

	if err != nil {
		log.Printf("can't generate url: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	str, err := url.JoinPath(conf.FlagBaseURL, "/", rndURL)

	if err != nil {
		log.Printf("can't join path %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

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

	res.Header().Set("Content-Type", "application/json")

	rndURL, err := generateURLAndSave(lenString, storage, bodyReq.URL)

	if errors.Is(err, helpers.ErrConflict) {
		res.WriteHeader(http.StatusConflict)
		str, err := url.JoinPath(conf.FlagBaseURL, "/", rndURL)

		if err != nil {
			log.Printf("can't join path %v", err)
			http.Error(res, "", http.StatusInternalServerError)
			return
		}
		_, err = res.Write([]byte(str))

		if err != nil {
			log.Printf("can't generate short url for original url %v", err)
			http.Error(res, "", http.StatusBadRequest)
			return
		}
		return
	}

	if err != nil {
		log.Printf("can't generate url: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	var bodyResp BodyResponse

	str, err := url.JoinPath(conf.FlagBaseURL, "/", rndURL)

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

	res.WriteHeader(http.StatusCreated)

	_, err = res.Write(resp)
	if err != nil {
		http.Error(res, "", http.StatusInternalServerError)
		return
	}
}

func GetPingHandler(res http.ResponseWriter, req *http.Request, storage storages.URLStorage) {
	pinger, ok := storage.(Pinger)
	if !ok {
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	err := pinger.CheckPing()
	if err != nil {
		log.Printf("failed to ping DB: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}

func BatchShortenHandler(res http.ResponseWriter, req *http.Request, storage storages.URLStorage, conf *config.Cfg) {
	if req.Body == http.NoBody {
		http.Error(res, "Empty String!", http.StatusBadRequest)
		return
	}

	bodyReq := []helpers.Incoming{}

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

	res.Header().Set("Content-Type", "application/json")

	batch, ok := storage.(helpers.BatchSaver)

	if !ok {
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	result, err := batch.Save(bodyReq, conf.FlagBaseURL)
	if err != nil {
		log.Printf("failed to save body: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(result)
	if err != nil {
		log.Printf("failed to marshal result: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusCreated)
	_, err = res.Write(data)
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

func generateURLAndSave(ln int, storage storages.URLStorage, originalURL string) (string, error) {
	rndString, err := generateRandom(ln, storage)

	if err != nil {
		return "", errors.New("failed to generate a random string")
	}
	err = storage.SaveURL(rndString, originalURL)

	if err != nil {
		var conflictErr *helpers.ConflictError
		if errors.As(err, &conflictErr) {
			rndString = conflictErr.ShortURL
		}

		return rndString, helpers.ErrConflict
	}
	return rndString, nil
}
