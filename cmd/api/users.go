package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bishop254/bursary/internal/mailer"
	"github.com/bishop254/bursary/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	// "github.com/gooogle/uuid"
)

type userKey string

const userCtx userKey = "user"
const studentCtx userKey = "student"
const userParamCtx userKey = "user_param"

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

	tmplVars := struct {
		Username string
		Link     string
	}{
		Username: user.Username,
		Link:     "htpp://lashkjsa",
	}

	err := a.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, tmplVars)
	if err != nil {
		if err := a.store.Users.RollBackNewUser(ctx, user.ID, hashToken); err != nil {
			a.internalServerError(w, r, err)
		}
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, user); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

func getUserFromCtx(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}

func getUserFromParamCtx(r *http.Request) *store.User {
	user, _ := r.Context().Value(userParamCtx).(*store.User)
	return user
}

func (a *application) getOneUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromParamCtx(r)

	if err := jsonResponse(w, http.StatusOK, user); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

// type FollowUserPayload struct {
// 	FollowerID int64 `json:"follower_id" validate:"required"`
// }

func (a *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)
	followedUser := getUserFromParamCtx(r)

	if followedUser.ID == user.ID {
		a.badRequestError(w, r, errors.New("cannot perform operation"))
		return
	}

	if err := a.store.Users.FollowUser(r.Context(), followedUser.ID, user.ID); err != nil {
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
	unfollowedUser := getUserFromParamCtx(r)

	if err := a.store.Users.UnfollowUser(r.Context(), unfollowedUser.ID, user.ID); err != nil {
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

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=3,max=71"`
}

func (a *application) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload LoginStudentPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	user, err := a.store.Users.GetOneByEmail(ctx, payload.Email)
	if err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := user.Password.CompareWithHash(payload.Password); err != nil {
		a.unauthorizedError(w, r, errors.New("invalid username/password"))
		return
	}

	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"iss":   "kcg",
		"aud":   "kcg",
		"exp":   time.Now().Add(a.config.auth.jwtAuth.exp).Unix(),
		"nbf":   time.Now().Unix(),
		"iat":   time.Now().Unix(),
		"jti":   uuid.New().String(),
	}

	token, err := a.authenticator.GenerateToken(claims)
	if err != nil {
		a.internalServerError(w, r, err)
		return
	}

	loginResp := struct {
		Token     string `json:"token"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		Blocked   bool   `json:"blocked"`
		CreatedAt string `json:"created_at"`
	}{
		Token:     token,
		Username:  user.Username,
		Email:     user.Email,
		Blocked:   user.Blocked,
		CreatedAt: user.CreatedAt,
	}

	if err := jsonResponse(w, http.StatusOK, loginResp); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}
