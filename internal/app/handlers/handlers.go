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

const marshalErrorTmp = "failed to marshal result: %v"         // marshalErrorTmp шаблон ошибки парсинга
const readBodyErrorTmp = "failed to read the request body: %v" // readBodyErrorTmp шаблон ошибки чтения тела запроса

// BodyRequested тело запроса на формирования короткой ссылки
type BodyRequested struct {
	// URL - url
	URL string `json:"url"`
}

// BodyResponse тело ответа с короткой сслыкой
type BodyResponse struct {
	// ShortURL = короткая ссылка
	ShortURL string `json:"result"`
}

// Pinger интерфейс для проверки состояния хранилища Postgres
type Pinger interface {
	CheckPing(ctx context.Context) error
}

// GetHandler запрос получения оригинальной ссылки по сокращенному URL
func GetHandler(_ context.Context, res http.ResponseWriter, req *http.Request, storage storages.URLStorage) {
	id := chi.URLParam(req, "id")

	// Получаем оригинальную ссылку из хранилища
	originalURL, err := storage.GetByID(req.Context(), id)

	if err != nil {
		var isDeletedErr *helpers.ConflictError
		if errors.As(err, &isDeletedErr) {
			res.WriteHeader(http.StatusGone)
			return
		}
		http.Error(res, "Not found", http.StatusNotFound)
		return
	}

	http.Redirect(res, req, originalURL, http.StatusTemporaryRedirect)
}

// PostHandler запрос на создание короткой ссылки для URL, Content-type: text/plain
func PostHandler(
	_ context.Context,
	res http.ResponseWriter,
	req *http.Request,
	storage storages.URLStorage,
	conf *config.Cfg,
	logger *zap.SugaredLogger,
) {
	if req.Body == http.NoBody {
		http.Error(res, "Empty String!", http.StatusBadRequest)
		return
	}

	u, err := io.ReadAll(req.Body)

	if err != nil {
		logger.Errorf(readBodyErrorTmp, err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	setHeader(res, "text/plain")

	// генерируем короткую ссылку и сохраняем
	rndURL, err := generateURLAndSave(req.Context(), storage, string(u))

	// обработка нестандратного поведения при сохранении, когда такой URL уже сущетсвует
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

// PostShortenHandler получение короткой ссылки для URL, тело в виде JSON
func PostShortenHandler(
	_ context.Context,
	res http.ResponseWriter,
	req *http.Request,
	storage storages.URLStorage,
	conf *config.Cfg,
	logger *zap.SugaredLogger,
) {
	if req.Body == http.NoBody {
		http.Error(res, "Empty String!", http.StatusBadRequest)
		return
	}

	bodyReq := BodyRequested{}

	body, err := io.ReadAll(req.Body)

	if err != nil {
		logger.Errorf(readBodyErrorTmp, err)
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

	rndURL, err := generateURLAndSave(req.Context(), storage, bodyReq.URL)

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

// GetPingHandler проверка подключения к хранилищу
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

// BatchShortenHandler запрос на массовое сохранение списка ссылок
func BatchShortenHandler(
	_ context.Context,
	res http.ResponseWriter,
	req *http.Request,
	storage storages.URLStorage,
	conf *config.Cfg,
	logger *zap.SugaredLogger,
) {
	if req.Body == http.NoBody {
		http.Error(res, "Empty String!", http.StatusBadRequest)
		return
	}

	var bodyReq []storages.Incoming

	body, err := io.ReadAll(req.Body)

	if err != nil {
		logger.Errorf(readBodyErrorTmp, err)
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

	// сохраняем полученные данные в хранилище
	result, err := storage.LoadURLs(req.Context(), bodyReq, conf.FlagBaseURL)

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

// GetUserUrls запрос на получение списка сохраненных пользователем URL
func GetUserUrls(
	_ context.Context,
	res http.ResponseWriter,
	req *http.Request,
	storage storages.URLStorage,
	conf *config.Cfg,
	logger *zap.SugaredLogger,
) {
	result, err := storage.GetUserURLs(req.Context(), conf.FlagBaseURL)

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

	logger.Infof("Marshalled data: %s", string(data))
	setHeader(res, "application/json")

	res.WriteHeader(http.StatusOK)
	_, err = res.Write(data)
	if err != nil {
		logger.Errorf("failed to write data: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}
}

// DeleteUserUrls запрос на мягкое удаление ссылок пользователя
func DeleteUserUrls(
	_ context.Context,
	res http.ResponseWriter,
	req *http.Request,
	storage storages.URLStorage,
	logger *zap.SugaredLogger,
) {
	if req.Body == http.NoBody {
		http.Error(res, "Empty Body!", http.StatusBadRequest)
		return
	}

	var bodyReq []string

	err := json.NewDecoder(req.Body).Decode(&bodyReq)

	if err != nil {
		logger.Errorf(readBodyErrorTmp, err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	err = storage.DeleteUserURLs(req.Context(), bodyReq, logger)

	if err != nil {
		logger.Errorf("failed to get delete URLs: %v", err)
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusAccepted)
}

func setHeader(res http.ResponseWriter, value string) {
	res.Header().Set("Content-Type", value)
}

func generateURLAndSave(
	ctx context.Context,
	storage storages.URLStorage,
	originalURL string,
) (string, error) {
	rndString, err := storage.SaveURL(ctx, originalURL)

	if err != nil {
		var conflictErr *helpers.ConflictError
		if errors.As(err, &conflictErr) {
			rndString = conflictErr.ShortURL
		}

		return rndString, helpers.ErrConflict
	}
	return rndString, nil
}
