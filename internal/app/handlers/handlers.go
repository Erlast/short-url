package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/helpers"
	"github.com/Erlast/short-url.git/internal/app/storages"
)

const lenString = 7
const marshalErrorTmp = "failed to marshal result: %v"

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
	conf *config.Cfg,
	logger *zap.SugaredLogger,
	user *storages.CurrentUser,
) {
	if req.Body == http.NoBody {
		http.Error(res, "Empty String!", http.StatusBadRequest)
		return
	}

	u, err := io.ReadAll(req.Body)

	if err != nil {
		logger.Errorf("failed to read the request body: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	setHeader(res, "text/plain")

	rndURL, err := generateURLAndSave(ctx, lenString, storage, string(u), user)

	if errors.Is(err, helpers.ErrConflict) {
		res.WriteHeader(http.StatusConflict)
		str, err := url.JoinPath(conf.FlagBaseURL, "/", rndURL)

		if err != nil {
			logger.Errorf("can't join path %v", err)
			http.Error(res, "", http.StatusInternalServerError)
			return
		}
		_, err = res.Write([]byte(str))

		if err != nil {
			http.Error(res, "can't generate short url for original url", http.StatusBadRequest)
			return
		}
		return
	}

	if err != nil {
		logger.Errorf("can't generate url: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	str, err := url.JoinPath(conf.FlagBaseURL, "/", rndURL)

	if err != nil {
		logger.Errorf("can't join path %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusCreated)

	_, err = res.Write([]byte(str))
	if err != nil {
		logger.Errorf("can't write body %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}
}

func PostShortenHandler(
	ctx context.Context,
	res http.ResponseWriter,
	req *http.Request,
	storage storages.URLStorage,
	conf *config.Cfg,
	logger *zap.SugaredLogger,
	user *storages.CurrentUser,
) {
	if req.Body == http.NoBody {
		http.Error(res, "Empty String!", http.StatusBadRequest)
		return
	}

	bodyReq := BodyRequested{}

	body, err := io.ReadAll(req.Body)

	if err != nil {
		logger.Errorf("failed to read the request body: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &bodyReq)

	if err != nil {
		logger.Errorf("failed to unmarshal body: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	setHeader(res, "application/json")

	rndURL, err := generateURLAndSave(ctx, lenString, storage, bodyReq.URL, user)

	if errors.Is(err, helpers.ErrConflict) {
		res.WriteHeader(http.StatusConflict)
		str, err := url.JoinPath(conf.FlagBaseURL, "/", rndURL)

		if err != nil {
			logger.Errorf("can't join path %v", err)
			http.Error(res, "", http.StatusInternalServerError)
			return
		}

		var bodyResp BodyResponse
		bodyResp.ShortURL = str

		resp, err := json.Marshal(bodyResp)

		if err != nil {
			logger.Errorf(marshalErrorTmp, err)
			http.Error(res, "", http.StatusInternalServerError)
			return
		}

		_, err = res.Write(resp)

		if err != nil {
			http.Error(res, "can't generate short url for original url", http.StatusBadRequest)
			return
		}
		return
	}

	if err != nil {
		logger.Errorf("can't generate url: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	var bodyResp BodyResponse

	str, err := url.JoinPath(conf.FlagBaseURL, "/", rndURL)

	if err != nil {
		logger.Errorf("can't join path: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	bodyResp.ShortURL = str

	resp, err := json.Marshal(bodyResp)

	if err != nil {
		logger.Errorf(marshalErrorTmp, err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusCreated)

	_, err = res.Write(resp)
	if err != nil {
		logger.Errorf("failed to write body: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}
}

func GetPingHandler(
	ctx context.Context,
	res http.ResponseWriter,
	storage storages.URLStorage,
	logger *zap.SugaredLogger,
) {
	pinger, ok := storage.(Pinger)
	if !ok {
		logger.Errorf("failed to ping DB: %v", ok)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	err := pinger.CheckPing(ctx)
	if err != nil {
		logger.Errorf("failed to ping DB: %v", err)
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
	conf *config.Cfg,
	logger *zap.SugaredLogger,
	user *storages.CurrentUser,
) {
	if req.Body == http.NoBody {
		http.Error(res, "Empty String!", http.StatusBadRequest)
		return
	}

	var bodyReq []storages.Incoming

	body, err := io.ReadAll(req.Body)

	if err != nil {
		logger.Errorf("failed to read the request body: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &bodyReq)

	if err != nil {
		logger.Errorf(marshalErrorTmp, err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}
	setHeader(res, "application/json")

	result, err := storage.LoadURLs(ctx, bodyReq, conf.FlagBaseURL, user)

	if err != nil {
		var conflictErr *helpers.ConflictError
		if errors.As(err, &conflictErr) {
			res.WriteHeader(http.StatusConflict)
			return
		}

		logger.Errorf("failed to save body: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(result)
	if err != nil {
		logger.Errorf(marshalErrorTmp, err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusCreated)
	_, err = res.Write(data)
	if err != nil {
		logger.Errorf("failed to write data: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}
}

func GetUserUrls(
	ctx context.Context,
	res http.ResponseWriter,
	storage storages.URLStorage,
	conf *config.Cfg,
	logger *zap.SugaredLogger,
	user *storages.CurrentUser,
) {
	result, err := storage.GetUserURLs(ctx, conf.FlagBaseURL, user)

	if err != nil {
		logger.Errorf("failed to get user URLs: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	if result == nil {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	data, err := json.Marshal(result)
	if err != nil {
		logger.Errorf(marshalErrorTmp, err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	setHeader(res, "application/json")

	res.WriteHeader(http.StatusCreated)
	_, err = res.Write(data)
	if err != nil {
		logger.Errorf("failed to write data: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}
}

func setHeader(res http.ResponseWriter, value string) {
	res.Header().Set("Content-Type", value)
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

func generateURLAndSave(
	ctx context.Context,
	ln int,
	storage storages.URLStorage,
	originalURL string,
	user *storages.CurrentUser,
) (string, error) {
	rndString, err := generateRandom(ctx, ln, storage)

	if err != nil {
		return "", errors.New("failed to generate a random string")
	}
	err = storage.SaveURL(ctx, rndString, originalURL, user)

	if err != nil {
		var conflictErr *helpers.ConflictError
		if errors.As(err, &conflictErr) {
			rndString = conflictErr.ShortURL
		}

		return rndString, helpers.ErrConflict
	}
	return rndString, nil
}
