package app

import (
	"github.com/Erlast/short-url.git/internal/config"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/url"
)

type Settings struct {
	Storage map[string]string
	Conf    config.Cfg
}

var settings Settings

func Init(s Settings) {
	settings = s
}

func GetHandler(res http.ResponseWriter, req *http.Request) {

	id := chi.URLParam(req, "id")

	originalURL, ok := settings.Storage[id]

	if !ok {
		http.Error(res, "Not Found!", http.StatusNotFound)
		return
	}

	http.Redirect(res, req, originalURL, http.StatusTemporaryRedirect)
}

func PostHandler(res http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()

	if req.Body == http.NoBody {
		http.Error(res, "Empty String!", http.StatusBadRequest)
		return
	}

	u, err := io.ReadAll(req.Body)

	if err != nil {
		http.Error(res, "Something went wrong!", http.StatusBadRequest)
	}

	rndString := GenerateRandom(7)

	settings.Storage[rndString] = string(u)

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
	rndString := RandomString(ln)

	if _, ok := settings.Storage[rndString]; ok {
		GenerateRandom(ln)
	}

	return rndString
}
