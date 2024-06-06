package handlers

import (
	"context"
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
	CheckPing(ctx context.Context) error
}

func GetHandler(ctx context.Context, res http.ResponseWriter, req *http.Request, storage storages.URLStorage) {
	id := chi.URLParam(req, "id")

	originalURL, ok := storage.GetByID(ctx, id)

	if ok != nil {
		http.Error(res, "Not found", http.StatusNotFound)
		return
	}

	http.Redirect(res, req, originalURL, http.StatusTemporaryRedirect)
}

func PostHandler(
	ctx context.Context,
	res http.ResponseWriter,
	req *http.Request,
	storage storages.URLStorage,
	conf *config.Cfg) {
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

	rndURL, err := generateURLAndSave(ctx, lenString, storage, string(u))

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

func PostShortenHandler(
	ctx context.Context,
	res http.ResponseWriter,
	req *http.Request,
	storage storages.URLStorage,
	conf *config.Cfg) {
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

	rndURL, err := generateURLAndSave(ctx, lenString, storage, bodyReq.URL)

	if errors.Is(err, helpers.ErrConflict) {
		res.WriteHeader(http.StatusConflict)
		str, err := url.JoinPath(conf.FlagBaseURL, "/", rndURL)

		if err != nil {
			log.Printf("can't join path %v", err)
			http.Error(res, "", http.StatusInternalServerError)
			return
		}

		var bodyResp BodyResponse
		bodyResp.ShortURL = str

		resp, err := json.Marshal(bodyResp)

		if err != nil {
			log.Printf("failed to marshal result: %v", err)
			http.Error(res, "", http.StatusInternalServerError)
			return
		}

		_, err = res.Write(resp)

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
		log.Printf("failed to write body: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}
}

func GetPingHandler(ctx context.Context, res http.ResponseWriter, _ *http.Request, storage storages.URLStorage) {
	pinger, ok := storage.(Pinger)
	if !ok {
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	err := pinger.CheckPing(ctx)
	if err != nil {
		log.Printf("failed to ping DB: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}

func BatchShortenHandler(
	ctx context.Context,
	res http.ResponseWriter,
	req *http.Request,
	storage storages.URLStorage,
	conf *config.Cfg) {
	if req.Body == http.NoBody {
		http.Error(res, "Empty String!", http.StatusBadRequest)
		return
	}

	var bodyReq []storages.Incoming

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

	result, err := storage.LoadURLs(ctx, bodyReq, conf.FlagBaseURL)

	if err != nil {
		var conflictErr *helpers.ConflictError
		if errors.As(err, &conflictErr) {
			res.WriteHeader(http.StatusConflict)
			log.Printf("some urls not original: %v", err)
			return
		}

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
		log.Printf("failed to write data: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}
}

func generateRandom(ctx context.Context, ln int, storage storages.URLStorage) (string, error) {
	for range 3 {
		rndString := helpers.RandomString(ln)

		if !storage.IsExists(ctx, rndString) {
			return rndString, nil
		}
	}
	return "", errors.New("failed to generate a unique string")
}

func generateURLAndSave(ctx context.Context, ln int, storage storages.URLStorage, originalURL string) (string, error) {
	rndString, err := generateRandom(ctx, ln, storage)

	if err != nil {
		return "", errors.New("failed to generate a random string")
	}
	err = storage.SaveURL(ctx, rndString, originalURL)

	if err != nil {
		var conflictErr *helpers.ConflictError
		if errors.As(err, &conflictErr) {
			rndString = conflictErr.ShortURL
		}

		return rndString, helpers.ErrConflict
	}
	return rndString, nil
}
