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
)

type RegisterStudentPayload struct {
	Firstname  string `json:"firstname" validate:"required"`
	Middlename string `json:"middlename"`
	Lastname   string `json:"lastname" validate:"required"`
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=8"`
}

func (a *application) registerStudentHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterStudentPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	student := &store.Student{
		Firstname:  payload.Firstname,
		Middlename: payload.Middlename,
		Lastname:   payload.Lastname,
		Email:      payload.Email,
	}

	if err := student.Password.Hashing(payload.Password); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	plainToken := uuid.New().String()
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	if err := a.store.Students.RegisterAndInvite(ctx, student, hashToken, time.Duration(a.config.mail.tokenExp)); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	// hashLink := fmt.Sprintf("http://localhost:8080/v8/students/activate/%s", hashToken)
	hashLink := fmt.Sprintf("http://localhost:4200/auth/activate/%s", hashToken)

	tmplVars := struct {
		Username string
		Link     string
	}{
		Username: student.Firstname,
		Link:     hashLink,
	}

	err := a.mailer.Send(mailer.UserWelcomeTemplate, student.Firstname, student.Email, tmplVars)
	if err != nil {
		if err := a.store.Students.RollBackNewStudent(ctx, student.ID, hashToken); err != nil {
			a.internalServerError(w, r, err)
		}
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, student); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

func (a *application) activateStudentHandler(w http.ResponseWriter, r *http.Request) {
	tokenParam := chi.URLParam(r, "token")

	ctx := r.Context()

	if err := a.store.Students.Activate(ctx, tokenParam); err != nil {
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

type LoginStudentPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=3,max=71"`
}

func (a *application) loginStudentHandler(w http.ResponseWriter, r *http.Request) {
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

	student, err := a.store.Students.GetOneByEmail(ctx, payload.Email)
	if err != nil {
		a.unauthorizedError(w, r, err)
		return
	}

	if err := student.Password.CompareWithHash(payload.Password); err != nil {
		a.unauthorizedError(w, r, errors.New("invalid username/password"))
		return
	}

	claims := jwt.MapClaims{
		"sub":   student.ID,
		"email": student.Email,
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
		Firstname: student.Firstname,
		Email:     student.Email,
		UID:       student.ID,
		Blocked:   student.Blocked,
		CreatedAt: student.CreatedAt,
	}

	if err := jsonResponse(w, http.StatusOK, loginResp); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

func getStudentFromCtx(r *http.Request) *store.Student {
	student, _ := r.Context().Value(studentCtx).(*store.Student)
	return student
}

// Personal
func (a *application) getStudentPersonalHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	ctx := r.Context()

	studentPersonalData, err := a.store.Students.GetStudentPersonalByID(ctx, student.ID)
	if err != nil {
		a.notFoundError(w, r, err)
		return
	}

	student.Personal = store.StudentPersonal(*studentPersonalData)

	if err := jsonResponse(w, http.StatusOK, student); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type StudentPersonalPayload struct {
	Dob              string `json:"dob" validate:"required"`
	Gender           string `json:"gender" validate:"required"`
	Citizenship      string `json:"citizenship" validate:"required"`
	Religion         string `json:"religion" validate:"required"`
	ParentalStatus   string `json:"parental_status" validate:"required"`
	BirthCertNo      string `json:"birth_cert_no" validate:"required"`
	BirthTown        string `json:"birth_town,omitempty"`
	BirthCounty      string `json:"birth_county" validate:"required"`
	BirthSubCounty   string `json:"birth_sub_county" validate:"required"`
	Ward             string `json:"ward" validate:"required"`
	VotersCardNo     string `json:"voters_card_no,omitempty"`
	Residence        string `json:"residence" validate:"required"`
	IDNumber         *int64 `json:"id_number,omitempty"`
	Phone            int64  `json:"phone" validate:"required"`
	KraPinNo         string `json:"kra_pin_no,omitempty"`
	PassportNo       string `json:"passport_no,omitempty"`
	SpecialNeed      bool   `json:"special_need"`
	SpecialNeedsType string `json:"special_needs_type" validate:"required"`
}

func (a *application) createStudentPersonalHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	var payload StudentPersonalPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	studentPersonalData := &store.StudentPersonal{
		Dob:              payload.Dob,
		Gender:           payload.Gender,
		Citizenship:      payload.Citizenship,
		Religion:         payload.Religion,
		ParentalStatus:   payload.ParentalStatus,
		BirthCertNo:      payload.BirthCertNo,
		BirthTown:        payload.BirthTown,
		BirthCounty:      payload.BirthCounty,
		BirthSubCounty:   payload.BirthSubCounty,
		Ward:             payload.Ward,
		VotersCardNo:     payload.VotersCardNo,
		Residence:        payload.Residence,
		IDNumber:         payload.IDNumber,
		Phone:            payload.Phone,
		KraPinNo:         payload.KraPinNo,
		PassportNo:       payload.PassportNo,
		SpecialNeed:      payload.SpecialNeed,
		SpecialNeedsType: payload.SpecialNeedsType,
		StudentID:        student.ID,
	}

	ctx := r.Context()

	if err := a.store.Students.CreateStudentPersonal(ctx, *studentPersonalData, student.ID); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, student); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type UpdateStudentPersonalPayload struct {
	ID               int64  `json:"id" validate:"required"`
	Dob              string `json:"dob" validate:"required"`
	Gender           string `json:"gender" validate:"required"`
	Citizenship      string `json:"citizenship" validate:"required"`
	Religion         string `json:"religion" validate:"required"`
	ParentalStatus   string `json:"parental_status" validate:"required"`
	BirthCertNo      string `json:"birth_cert_no" validate:"required"`
	BirthTown        string `json:"birth_town,omitempty"`
	BirthCounty      string `json:"birth_county" validate:"required"`
	BirthSubCounty   string `json:"birth_sub_county" validate:"required"`
	Ward             string `json:"ward" validate:"required"`
	VotersCardNo     string `json:"voters_card_no,omitempty"`
	Residence        string `json:"residence" validate:"required"`
	IDNumber         *int64 `json:"id_number,omitempty"`
	Phone            int64  `json:"phone" validate:"required"`
	KraPinNo         string `json:"kra_pin_no,omitempty"`
	PassportNo       string `json:"passport_no,omitempty"`
	SpecialNeed      bool   `json:"special_need"`
	SpecialNeedsType string `json:"special_needs_type" validate:"required"`
}

func (a *application) updateStudentPersonalHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	var payload UpdateStudentPersonalPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	studentPersonalData := &store.StudentPersonal{
		ID:               payload.ID,
		Dob:              payload.Dob,
		Gender:           payload.Gender,
		Citizenship:      payload.Citizenship,
		Religion:         payload.Religion,
		ParentalStatus:   payload.ParentalStatus,
		BirthCertNo:      payload.BirthCertNo,
		BirthTown:        payload.BirthTown,
		BirthCounty:      payload.BirthCounty,
		BirthSubCounty:   payload.BirthSubCounty,
		Ward:             payload.Ward,
		VotersCardNo:     payload.VotersCardNo,
		Residence:        payload.Residence,
		IDNumber:         payload.IDNumber,
		Phone:            payload.Phone,
		KraPinNo:         payload.KraPinNo,
		PassportNo:       payload.PassportNo,
		SpecialNeed:      payload.SpecialNeed,
		SpecialNeedsType: payload.SpecialNeedsType,
		StudentID:        student.ID,
	}

	ctx := r.Context()

	if err := a.store.Students.UpdateStudentPersonal(ctx, *studentPersonalData, student.ID); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, student); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

// Institution
func (a *application) getStudentInstitutionHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	ctx := r.Context()

	studentInstitutionData, err := a.store.Students.GetStudentInstitutionByID(ctx, student.ID)
	if err != nil {
		a.notFoundError(w, r, err)
		return
	}

	student.Institution = store.StudentInstitution(*studentInstitutionData)

	if err := jsonResponse(w, http.StatusOK, studentInstitutionData); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type StudentInstitutionPayload struct {
	InstName       string `json:"inst_name" validate:"required"`
	InstType       string `json:"inst_type" validate:"required"`
	Category       string `json:"category,omitempty"`
	Telephone      int64  `json:"telephone" validate:"required"`
	Email          string `json:"email,omitempty"`
	Address        string `json:"address" validate:"required"`
	InstCounty     string `json:"inst_county" validate:"required"`
	InstSubCounty  string `json:"inst_sub_county" validate:"required"`
	InstWard       string `json:"inst_ward,omitempty"`
	PrincipalName  string `json:"principal_name" validate:"required"`
	YearJoined     int64  `json:"year_joined" validate:"required"`
	CurrClassLevel string `json:"curr_class_level" validate:"required"`
	AdmNo          string `json:"adm_no" validate:"required"`
	BankName       string `json:"bank_name" validate:"required"`
	BankBranch     string `json:"bank_branch" validate:"required"`
	BankAccName    string `json:"bank_acc_name" validate:"required"`
	BankAccNo      int64  `json:"bank_acc_no" validate:"required"`
}

func (a *application) createStudentInstitutionHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	var payload StudentInstitutionPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	studentInstitutionData := &store.StudentInstitution{
		InstName:       payload.InstName,
		InstType:       payload.InstType,
		Category:       payload.Category,
		Telephone:      payload.Telephone,
		Email:          payload.Email,
		Address:        payload.Address,
		InstCounty:     payload.InstCounty,
		InstSubCounty:  payload.InstSubCounty,
		InstWard:       payload.InstWard,
		PrincipalName:  payload.PrincipalName,
		YearJoined:     payload.YearJoined,
		CurrClassLevel: payload.CurrClassLevel,
		AdmNo:          payload.AdmNo,
		StudentID:      student.ID,
		BankName:       payload.BankName,
		BankBranch:     payload.BankBranch,
		BankAccName:    payload.BankAccName,
		BankAccNo:      payload.BankAccNo,
	}

	ctx := r.Context()

	if err := a.store.Students.CreateStudentInstitution(ctx, *studentInstitutionData, student.ID); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, student); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type UpdateStudentInstitutionPayload struct {
	ID             int64  `json:"id" validate:"required"`
	InstName       string `json:"inst_name" validate:"required"`
	InstType       string `json:"inst_type" validate:"required"`
	Category       string `json:"category,omitempty"`
	Telephone      int64  `json:"telephone" validate:"required"`
	Email          string `json:"email,omitempty"`
	Address        string `json:"address" validate:"required"`
	InstCounty     string `json:"inst_county" validate:"required"`
	InstSubCounty  string `json:"inst_sub_county" validate:"required"`
	InstWard       string `json:"inst_ward,omitempty"`
	PrincipalName  string `json:"principal_name" validate:"required"`
	YearJoined     int64  `json:"year_joined" validate:"required"`
	CurrClassLevel string `json:"curr_class_level" validate:"required"`
	AdmNo          string `json:"adm_no" validate:"required"`
	BankName       string `json:"bank_name" validate:"required"`
	BankBranch     string `json:"bank_branch" validate:"required"`
	BankAccName    string `json:"bank_acc_name" validate:"required"`
	BankAccNo      int64  `json:"bank_acc_no" validate:"required"`
}

func (a *application) updateStudentInstitutionHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	var payload UpdateStudentInstitutionPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	studentInstitutionData := &store.StudentInstitution{
		ID:             payload.ID,
		InstName:       payload.InstName,
		InstType:       payload.InstType,
		Category:       payload.Category,
		Telephone:      payload.Telephone,
		Email:          payload.Email,
		Address:        payload.Address,
		InstCounty:     payload.InstCounty,
		InstSubCounty:  payload.InstSubCounty,
		InstWard:       payload.InstWard,
		PrincipalName:  payload.PrincipalName,
		YearJoined:     payload.YearJoined,
		CurrClassLevel: payload.CurrClassLevel,
		AdmNo:          payload.AdmNo,
		StudentID:      student.ID,
		BankName:       payload.BankName,
		BankBranch:     payload.BankBranch,
		BankAccName:    payload.BankAccName,
		BankAccNo:      payload.BankAccNo,
	}

	ctx := r.Context()

	if err := a.store.Students.UpdateStudentInstitution(ctx, *studentInstitutionData, student.ID); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, studentInstitutionData); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}
