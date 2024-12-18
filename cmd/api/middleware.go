package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func (a *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//receive base64 encoded string from Authorization Header
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				a.unauthorizedError(w, r, errors.New("authorization header missing"))
				return
			}

			authHeaderSlice := strings.Split(authHeader, " ")
			if len(authHeaderSlice) != 2 || authHeaderSlice[0] != "Basic" {
				a.unauthorizedError(w, r, errors.New("authorization header malformed"))
				return
			}

			decodedStr, err := base64.StdEncoding.DecodeString(authHeaderSlice[1])
			if err != nil {
				a.unauthorizedError(w, r, errors.New("authorization header malformed"))
				return
			}

			username := a.config.auth.basicAuth.username
			password := a.config.auth.basicAuth.password
			decodedPart := strings.SplitN(string(decodedStr), ":", 2)

			fmt.Println(decodedPart)
			fmt.Println(decodedPart[0])
			fmt.Println(decodedPart[1])
			if len(decodedPart) != 2 || decodedPart[0] != username || decodedPart[1] != password {
				a.unauthorizedError(w, r, errors.New("authorization header invalid"))
				return
			}

			next.ServeHTTP(w, r)
		})

	}
}
