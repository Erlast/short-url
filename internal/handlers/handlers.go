package handlers

import (
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"github.com/Erlast/short-url.git/internal/config"
	"github.com/Erlast/short-url.git/internal/helpers"
	"github.com/Erlast/short-url.git/internal/storages"
)

type Settings struct {
	Storage storages.Storage
	Conf    config.Cfg
}

var settings Settings

func Init(s storages.Storage, c config.Cfg) {
	settings = Settings{
		Storage: s,
		Conf:    c,
	}
}

func GetHandler(res http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	originalURL, ok := settings.Storage.Urls[id]

	if !ok {
		http.Error(res, "Not Found!", http.StatusNotFound)
		return
	}

	http.Redirect(res, req, originalURL, http.StatusTemporaryRedirect)
}

func PostHandler(res http.ResponseWriter, req *http.Request) {
	if req.Body == http.NoBody {
		http.Error(res, "Empty String!", http.StatusBadRequest)
		return
	}

	u, err := io.ReadAll(req.Body)

	if err != nil {
		http.Error(res, "Something went wrong!", http.StatusBadRequest)
		return
	}

	err = req.Body.Close()

	if err != nil {
		http.Error(res, "Empty String!", http.StatusInternalServerError)
		return
	}

	const lenString = 7

	rndString := GenerateRandom(lenString)

	settings.Storage.Urls[rndString] = string(u)

	str, err := url.JoinPath(settings.Conf.FlagBaseURL, "/", rndString)

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

func GenerateRandom(ln int) string {
	rndString := helpers.RandomString(ln)

	if _, ok := settings.Storage.Urls[rndString]; ok {
		GenerateRandom(ln)
	}

	return rndString
}
