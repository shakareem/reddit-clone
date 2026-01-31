package handlers

import (
	"context"
	"fmt"
	"net/http"
	"redditclone/internal/storage"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	ID   string `json:"id"`
	Name string `json:"username"`
}

type Claims struct {
	User UserClaims `json:"user"`
	jwt.RegisteredClaims
}

const jwtSecret = "abc" // tmp

func generateJWT(user storage.User) (string, error) {
	claims := Claims{
		User: UserClaims{
			ID:   user.ID,
			Name: user.Name,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func parseJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("bad sign method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}

func withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		inToken := ""
		if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
			inToken = after
		}
		claims, err := parseJWT(inToken)
		if err != nil {
			http.Error(w, `{"message":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), USER, claims.User)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
