package middlewares

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/helpers"
)

// Claims содержимое jwt токена
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const tokenExp = time.Hour * 3          // tokenExp время жизни токена
const accessDeniedErr = "Access denied" // accessDeniedErr шаблон ошибки Доступ запрещен

// AuthMiddleware функция установки jwt токенов в заголовок http запроса и в cookie, если их нет
func AuthMiddleware(h http.Handler, logger *zap.SugaredLogger, cfg *config.Cfg) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		token, err := req.Cookie("token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				newToken, err := buildJWTString(cfg)
				if err != nil {
					logger.Errorln("Failed to build token", err)
					http.Error(resp, "", http.StatusInternalServerError)
					return
				}
				resp.Header().Set("Authorization", newToken)
				http.SetCookie(resp, &http.Cookie{Name: "token", Value: newToken})

				userID := getUserID(newToken, logger, cfg)
				if userID == "" {
					http.Error(resp, accessDeniedErr, http.StatusUnauthorized)
					return
				}

				ctx := context.WithValue(req.Context(), helpers.UserID, userID)

				req = req.WithContext(ctx)

				h.ServeHTTP(resp, req)
				return
			}
			logger.Warnw("Failed to get token from cookie", "error", err)
			http.Error(resp, "", http.StatusInternalServerError)
			return
		}

		userID := getUserID(token.Value, logger, cfg)
		if userID == "" {
			http.Error(resp, accessDeniedErr, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(req.Context(), helpers.UserID, userID)

		req = req.WithContext(ctx)

		h.ServeHTTP(resp, req)
	})
}

// CheckAuthMiddleware функция проверки jwt токенов из заголовков http запроса
func CheckAuthMiddleware(h http.Handler, logger *zap.SugaredLogger) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		authorization := req.Header.Get("Authorization")
		if authorization == "" {
			http.Error(resp, accessDeniedErr, http.StatusUnauthorized)
			return
		}
		token, err := req.Cookie("token")
		if err != nil {
			logger.Errorln("Failed to get token from cookie", err)
			http.Error(resp, "", http.StatusInternalServerError)
			return
		}

		if authorization != token.Value {
			http.Error(resp, accessDeniedErr, http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(resp, req)
	})
}

func buildJWTString(cfg *config.Cfg) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: uuid.NewString(),
	})

	tokenString, err := token.SignedString([]byte(cfg.SecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func getUserID(tokenString string, logger *zap.SugaredLogger, cfg *config.Cfg) string {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(cfg.SecretKey), nil
		})
	if err != nil {
		return ""
	}

	if !token.Valid {
		logger.Errorln("Token is not valid")
		return ""
	}

	return claims.UserID
}
