package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bishop254/bursary/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	// "github.com/gooogle/uuid"
)

type userKey string

const userCtx userKey = "user"

type CreateUserPayload struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (a *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	if err := user.Password.Hashing(payload.Password); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	plainToken := uuid.New().String()
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	if err := a.store.Users.CreateAndInvite(ctx, user, hashToken, time.Duration(a.config.mail.tokenExp)); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	//send email to the user

	if err := jsonResponse(w, http.StatusCreated, user); err != nil {
		a.internalServerError(w, r, err)
		return
	}
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

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromCtx(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}

func (a *application) getOneUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	if err := jsonResponse(w, http.StatusOK, user); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type FollowUserPayload struct {
	FollowerID int64 `json:"follower_id" validate:"required"`
}

func (a *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	var payload FollowUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if payload.FollowerID == user.ID {
		a.badRequestError(w, r, errors.New("cannot perform operation"))
		return
	}

	if err := a.store.Users.FollowUser(r.Context(), user.ID, payload.FollowerID); err != nil {
		fmt.Println(err)
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, nil); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

func (a *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	var payload FollowUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if payload.FollowerID == user.ID {
		a.badRequestError(w, r, errors.New("cannot perform operation"))
		return
	}

	if err := a.store.Users.UnfollowUser(r.Context(), user.ID, payload.FollowerID); err != nil {
		fmt.Println(err)
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, nil); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

func (a *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	tokenParam := chi.URLParam(r, "token")

	ctx := r.Context()

	if err := a.store.Users.Activate(ctx, tokenParam); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusNoContent, ""); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}
