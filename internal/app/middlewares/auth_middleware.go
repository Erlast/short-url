package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Erlast/short-url.git/internal/app/storages"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const TokenExp = time.Hour * 3
const SecretKey = "supersecretkey"

func AuthMiddleware(h http.Handler, logger *zap.SugaredLogger, user *storages.CurrentUser) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		token, err := req.Cookie("token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				newToken, err := buildJWTString()
				if err != nil {
					logger.Warnw("Failed to build token", "error", err)
					http.Error(resp, "", http.StatusInternalServerError)
					return
				}
				http.SetCookie(resp, &http.Cookie{Name: "token", Value: newToken})

				userID := getUserID(newToken, logger)
				if userID == "" {
					http.Error(resp, "Access denied", http.StatusUnauthorized)
					return
				}
				user.UserID = userID
				h.ServeHTTP(resp, req)
				return
			}
			logger.Warnw("Failed to get token from cookie", "error", err)
			http.Error(resp, "", http.StatusInternalServerError)
			return
		}

		userID := getUserID(token.Value, logger)
		if userID == "" {
			http.Error(resp, "Access denied", http.StatusUnauthorized)
			return
		}

		user.UserID = userID

		h.ServeHTTP(resp, req)
	})
}

func buildJWTString() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: uuid.NewString(),
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func getUserID(tokenString string, logger *zap.SugaredLogger) string {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SecretKey), nil
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
