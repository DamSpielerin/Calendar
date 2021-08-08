package server

import (
	"calendar/storage"
	"calendar/user"
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"time"
)

func PanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("panicMiddleware", r.URL.Path)
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("recovered", err)
				http.Error(w, "Internal server error", 500)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, "error: no cookie", http.StatusUnauthorized)
				return
			}
			http.Error(w, "error: wrong cookie", http.StatusBadRequest)
			return
		}
		tknStr := c.Value
		claims := &user.Claims{}

		tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
			return user.JwtKey, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				http.Error(w, "error: jwt signature is invalid", http.StatusUnauthorized)
				return
			}
			http.Error(w, "error: signature is invalid", http.StatusBadRequest)
			return
		}

		if !tkn.Valid {
			http.Error(w, "error: jwt is invalid", http.StatusUnauthorized)
			return
		}

		userEntity, ok := storage.Users.GetUserByLogin(claims.Username)
		if !ok {
			http.Error(w, "error: user not found", http.StatusNotFound)
		}

		if userEntity.Timezone != claims.Timezone {
			claims.Timezone = userEntity.Timezone

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString(user.JwtKey)
			if err != nil {
				http.Error(w, "error: can't update user's timezone", http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:    "token",
				Value:   tokenString,
				Expires: time.Unix(claims.StandardClaims.ExpiresAt, 0),
			})
		}
		ctx := context.WithValue(r.Context(), "timezone", userEntity.Timezone)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
