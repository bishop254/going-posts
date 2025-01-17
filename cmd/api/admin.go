package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bishop254/bursary/internal/mailer"
	"github.com/bishop254/bursary/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type LoginAdminPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=3,max=71"`
}

func (a *application) loginAdminHandler(w http.ResponseWriter, r *http.Request) {
	var payload LoginAdminPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	admin, err := a.store.Admins.GetOneByEmail(ctx, payload.Email)
	if err != nil {
		a.unauthorizedError(w, r, errors.New("invalid username/password"))
		return
	}

	if err := admin.Password.CompareWithHash(payload.Password); err != nil {
		a.unauthorizedError(w, r, errors.New("invalid username/password"))
		return
	}

	claims := jwt.MapClaims{
		"sub":   admin.ID,
		"email": admin.Email,
		"iss":   "migBurs",
		"aud":   "migBurs",
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
		Firstname string `json:"firstname"`
		Email     string `json:"email"`
		UID       int64  `json:"uid"`
		Blocked   bool   `json:"blocked"`
		CreatedAt string `json:"created_at"`
	}{
		Token:     token,
		Firstname: admin.Firstname,
		Email:     admin.Email,
		UID:       admin.ID,
		Blocked:   admin.Blocked,
		CreatedAt: admin.CreatedAt,
	}

	if err := jsonResponse(w, http.StatusOK, loginResp); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type RegisterAdminPayload struct {
	Firstname  string  `json:"firstname" validate:"required"`
	Middlename *string `json:"middlename"`
	Lastname   string  `json:"lastname" validate:"required"`
	Email      string  `json:"email" validate:"required,email"`
	Password   string  `json:"password" validate:"required,min=8"`
	Role       int64   `json:"role" validate:"required"`
	RoleCode   string  `json:"role_code" validate:"required"`
}

func (a *application) registerAdminHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterAdminPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	admin := &store.Admin{
		Firstname:  payload.Firstname,
		Middlename: payload.Middlename,
		Lastname:   payload.Lastname,
		Email:      payload.Email,
	}

	if err := admin.Password.Hashing(payload.Password); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	plainToken := uuid.New().String()
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	if err := a.store.Admins.RegisterAndInvite(ctx, admin, hashToken, time.Duration(a.config.mail.tokenExp)); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	hashLink := fmt.Sprintf("http://localhost:4300/auth/activate/%s", hashToken)

	tmplVars := struct {
		Username string
		Link     string
	}{
		Username: admin.Firstname,
		Link:     hashLink,
	}

	err := a.mailer.Send(mailer.UserWelcomeTemplate, admin.Firstname, admin.Email, tmplVars)
	if err != nil {
		if err := a.store.Admins.RollBackNewAdmin(ctx, admin.ID, hashToken); err != nil {
			a.internalServerError(w, r, err)
		}
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, admin); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

func (a *application) activateAdminHandler(w http.ResponseWriter, r *http.Request) {
	tokenParam := chi.URLParam(r, "token")

	ctx := r.Context()

	if err := a.store.Admins.Activate(ctx, tokenParam); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	activateResp := struct {
		Token string `json:"token"`
	}{
		Token: tokenParam,
	}

	if err := jsonResponse(w, http.StatusAccepted, activateResp); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

func (a *application) getBursariesHandler(w http.ResponseWriter, r *http.Request) {
	bursaryQuery := &store.PaginatedFeedQuery{
		Limit:  10,
		Offset: 10,
		Sort:   "desc",
	}

	bursaryQuery, err := bursaryQuery.Parse(r)
	if err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(bursaryQuery); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	bursaryListing, err := a.store.Bursaries.GetBursariesAndCount(ctx, bursaryQuery)
	if err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusAccepted, bursaryListing); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type BursaryPayload struct {
	BursaryName      string   `json:"bursary_name" validate:"required"`
	Description      *string  `json:"description"`
	EndDate          string   `json:"end_date" validate:"required"`
	AmountAllocated  *float64 `json:"amount_allocated"`
	AmountPerStudent *float64 `json:"amount_per_student"`
	AllocationType   string   `json:"allocation_type" validate:"required,oneof=fixed variable"`
}

func (a *application) createBursaryHandler(w http.ResponseWriter, r *http.Request) {
	var payload BursaryPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	bursary := &store.Bursary{
		BursaryName:      payload.BursaryName,
		Description:      payload.Description,
		EndDate:          payload.EndDate,
		AmountAllocated:  payload.AmountAllocated,
		AmountPerStudent: payload.AmountPerStudent,
		AllocationType:   payload.AllocationType,
		CreatedAt:        time.Now().String(),
	}

	if err := a.store.Bursaries.CreateBursary(ctx, *bursary); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, bursary); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type UpdateBursaryPayload struct {
	ID               int64    `json:"id" validate:"required"`
	BursaryName      string   `json:"bursary_name" validate:"required"`
	Description      *string  `json:"description"`
	EndDate          string   `json:"end_date" validate:"required"`
	AmountAllocated  *float64 `json:"amount_allocated"`
	AmountPerStudent *float64 `json:"amount_per_student"`
	AllocationType   string   `json:"allocation_type" validate:"required,oneof=fixed variable"`
}

func (a *application) updateBursaryHandler(w http.ResponseWriter, r *http.Request) {

	var payload UpdateBursaryPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	bursary := &store.Bursary{
		BursaryName:      payload.BursaryName,
		Description:      payload.Description,
		EndDate:          payload.EndDate,
		AmountAllocated:  payload.AmountAllocated,
		AmountPerStudent: payload.AmountPerStudent,
		AllocationType:   payload.AllocationType,
		CreatedAt:        time.Now().String(),
		ID:               payload.ID,
	}

	if err := a.store.Bursaries.UpdateBursary(ctx, *bursary); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusAccepted, bursary); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

func (a *application) getAdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	adminUserQuery := &store.PaginatedAdminUserQuery{
		Limit:  10,
		Offset: 10,
		Sort:   "desc",
	}

	adminUserQuery, err := adminUserQuery.ParseAdminUser(r)
	if err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(adminUserQuery); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	adminUsersListing, err := a.store.Admins.GetAdminUsers(ctx, adminUserQuery)
	if err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusAccepted, adminUsersListing); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

func (a *application) getRolesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	adminUsersListing, err := a.store.Admins.GetRoles(ctx)
	if err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusAccepted, adminUsersListing); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type CreateAdminPayload struct {
	Firstname  string  `json:"firstname" validate:"required"`
	Middlename *string `json:"middlename"`
	Lastname   string  `json:"lastname" validate:"required"`
	Email      string  `json:"email" validate:"required,email"`
	Role       int64   `json:"role" validate:"required"`
	RoleCode   string  `json:"role_code" validate:"required"`
}

func (a *application) createAdminUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateAdminPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	admin := &store.Admin{
		Firstname:  payload.Firstname,
		Middlename: payload.Middlename,
		Lastname:   payload.Lastname,
		Email:      payload.Email,
		Role: store.Role{
			ID: payload.Role,
		},
		RoleCode: &payload.RoleCode,
	}

	genPass := strings.Split(uuid.New().String(), "-")[0]

	if err := admin.Password.Hashing(genPass); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	plainToken := uuid.New().String()
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	if err := a.store.Admins.RegisterAndInvite(ctx, admin, hashToken, time.Duration(a.config.mail.tokenExp)); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	hashLink := fmt.Sprintf("http://localhost:4300/auth/activate/%s", hashToken)

	tmplVars := struct {
		Username string
		Link     string
		Pass     string
	}{
		Username: admin.Firstname,
		Link:     hashLink,
		Pass:     genPass,
	}

	err := a.mailer.Send(mailer.AdminUserWelcomeTemplate, admin.Firstname, admin.Email, tmplVars)
	if err != nil {
		if err := a.store.Admins.RollBackNewAdmin(ctx, admin.ID, hashToken); err != nil {
			a.internalServerError(w, r, err)
		}
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, admin); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}
