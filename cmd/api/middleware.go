package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bishop254/bursary/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

func (a *application) JWTAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				a.unauthorizedError(w, r, errors.New("authorization header missing"))
				return
			}

			authHeaderSlice := strings.Split(authHeader, " ")
			if len(authHeaderSlice) != 2 || authHeaderSlice[0] != "Bearer" {
				a.unauthorizedError(w, r, errors.New("authorization header malformed"))
				return
			}

			token := authHeaderSlice[1]

			tokenString, err := a.authenticator.ValidateToken(token)
			if err != nil {
				a.unauthorizedError(w, r, err)
				return
			}

			claims, _ := tokenString.Claims.(jwt.MapClaims)
			userId, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
			if err != nil {
				a.unauthorizedError(w, r, err)
				return
			}

			ctx := r.Context()

			user, err := a.store.Users.GetOne(ctx, userId)
			if err != nil {
				a.unauthorizedError(w, r, err)
				return
			}

			ctx = context.WithValue(ctx, userCtx, user)

			next.ServeHTTP(w, r.WithContext(ctx))
		})

	}
}

func (a *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			if len(decodedPart) != 2 || decodedPart[0] != username || decodedPart[1] != password {
				a.unauthorizedError(w, r, errors.New("authorization header invalid"))
				return
			}

			next.ServeHTTP(w, r)
		})

	}
}

func (a *application) AuthorizationCheck(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := getUserFromCtx(r)
		post := getPostFromCtx(r)

		if post.UserID == user.ID {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()

		isAllowed, err := a.isUserAuthorized(ctx, user, requiredRole)
		if err != nil {
			a.internalServerError(w, r, err)
			return
		}

		if !isAllowed {
			a.forbiddenError(w, r, errors.New("not authorized to access this resource"))
			return
		}

		next.ServeHTTP(w, r)
	})

}

func (a *application) isUserAuthorized(ctx context.Context, user *store.User, requiredRole string) (bool, error) {
	requiredRoleData, err := a.store.Roles.GetOneByName(ctx, requiredRole)
	if err != nil {
		return false, err
	}

	return user.Role.Level >= requiredRoleData.Level, nil
}

func (a *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		idParam := chi.URLParam(r, "userId")
		userId, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			a.internalServerError(w, r, err)
			return
		}

		user, err := a.store.Users.GetOne(ctx, userId)
		if err != nil {
			switch {
			case errors.Is(err, errors.New("user not found")):
				a.notFoundError(w, r, err)
				return
			default:
				a.internalServerError(w, r, err)
				return
			}
		}

		ctx = context.WithValue(ctx, userParamCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *application) postContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		idParam := chi.URLParam(r, "postId")
		postId, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			a.internalServerError(w, r, err)
			return
		}

		post, err := a.store.Posts.GetOne(ctx, postId)
		if err != nil {
			switch {
			case errors.Is(err, errors.New("post not found")):
				a.notFoundError(w, r, err)
				return
			default:
				a.internalServerError(w, r, err)
				return
			}
		}

		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
