package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/bishop254/bursary/internal/mailer"
	"github.com/bishop254/bursary/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
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

	// student.Personal = store.StudentPersonal(*studentPersonalData)

	if err := jsonResponse(w, http.StatusOK, studentPersonalData); err != nil {
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
	SpecialNeed      int64  `json:"special_need"`
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
	SpecialNeed      int64  `json:"special_need"`
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

// Sponsor
func (a *application) getStudentSponsorHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	ctx := r.Context()

	studentSponsorData, err := a.store.Students.GetStudentSponsorByID(ctx, student.ID)
	if err != nil {
		a.notFoundError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, studentSponsorData); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type StudentSponsorPayload struct {
	Name               string `json:"name" validate:"required"`
	SponsorshipType    string `json:"sponsorship_type" validate:"required"`
	SponsorshipNature  string `json:"sponsorship_nature" validate:"required"`
	Phone              int64  `json:"phone" validate:"required"`
	Email              string `json:"email,omitempty"`
	Address            string `json:"address,omitempty"`
	ContactPersonName  string `json:"contact_person_name,omitempty"`
	ContactPersonPhone int64  `json:"contact_person_phone,omitempty"`
}

func (a *application) createStudentSponsorHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	var payload StudentSponsorPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	studentSponsorData := &store.StudentSponsor{
		Name:               payload.Name,
		SponsorshipType:    payload.SponsorshipType,
		SponsorshipNature:  payload.SponsorshipNature,
		Phone:              payload.Phone,
		Email:              payload.Email,
		Address:            payload.Address,
		ContactPersonName:  payload.ContactPersonName,
		ContactPersonPhone: payload.ContactPersonPhone,
		StudentID:          student.ID,
	}

	ctx := r.Context()

	if err := a.store.Students.CreateStudentSponsor(ctx, *studentSponsorData, student.ID); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, student); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type UpdateStudentSponsorPayload struct {
	ID                 int64  `json:"id" validate:"required"`
	Name               string `json:"name" validate:"required"`
	SponsorshipType    string `json:"sponsorship_type" validate:"required"`
	SponsorshipNature  string `json:"sponsorship_nature" validate:"required"`
	Phone              int64  `json:"phone" validate:"required"`
	Email              string `json:"email,omitempty"`
	Address            string `json:"address" validate:"required"`
	ContactPersonName  string `json:"contact_person_name,omitempty"`
	ContactPersonPhone int64  `json:"contact_person_phone,omitempty"`
}

func (a *application) updateStudentSponsorHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	var payload UpdateStudentSponsorPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	studentSponsorData := &store.StudentSponsor{
		ID:                 payload.ID,
		Name:               payload.Name,
		SponsorshipType:    payload.SponsorshipType,
		SponsorshipNature:  payload.SponsorshipNature,
		Phone:              payload.Phone,
		Email:              payload.Email,
		Address:            payload.Address,
		ContactPersonName:  payload.ContactPersonName,
		ContactPersonPhone: payload.ContactPersonPhone,
		StudentID:          student.ID,
		UpdatedAt:          time.Now().String(),
	}

	ctx := r.Context()

	if err := a.store.Students.UpdateStudentSponsor(ctx, *studentSponsorData, student.ID); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, studentSponsorData); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

// Emergency
func (a *application) getStudentEmergencyHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	ctx := r.Context()

	studentEmergencyData, err := a.store.Students.GetStudentEmergencyByID(ctx, student.ID)
	if err != nil {
		a.notFoundError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, studentEmergencyData); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type StudentEmergencyPayload struct {
	Firstname    string `json:"firstname" validate:"required"`
	Middlename   string `json:"middlename,omitempty"`
	Lastname     string `json:"lastname" validate:"required"`
	Phone        int64  `json:"phone" validate:"required"`
	Email        string `json:"email,omitempty"`
	IdNumber     int64  `json:"id_number" validate:"required"`
	Occupation   string `json:"occupation,omitempty"`
	Relationship string `json:"relationship" validate:"required"`
	Residence    string `json:"residence" validate:"required"`
	Town         string `json:"town,omitempty"`
	WorkPlace    string `json:"work_place,omitempty"`
	WorkPhone    int64  `json:"work_phone,omitempty"`
	ProvidedBy   string `json:"provided_by,omitempty"`
}

func (a *application) createStudentEmergencyHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	var payload StudentEmergencyPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	studentEmergencyData := &store.StudentEmergency{
		Firstname:    payload.Firstname,
		Middlename:   payload.Middlename,
		Lastname:     payload.Lastname,
		Phone:        payload.Phone,
		Email:        payload.Email,
		IdNumber:     payload.IdNumber,
		Occupation:   payload.Occupation,
		Relationship: payload.Relationship,
		Residence:    payload.Residence,
		Town:         payload.Town,
		WorkPlace:    payload.WorkPlace,
		WorkPhone:    payload.WorkPhone,
		ProvidedBy:   payload.ProvidedBy,
		StudentID:    student.ID,
		UpdatedAt:    time.Now().String(),
	}

	ctx := r.Context()

	if err := a.store.Students.CreateStudentEmergency(ctx, *studentEmergencyData, student.ID); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, studentEmergencyData); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type UpdateStudentEmergencyPayload struct {
	ID           int64  `json:"id" validate:"required"`
	Firstname    string `json:"firstname" validate:"required"`
	Middlename   string `json:"middlename,omitempty"`
	Lastname     string `json:"lastname" validate:"required"`
	Phone        int64  `json:"phone" validate:"required"`
	Email        string `json:"email,omitempty"`
	IdNumber     int64  `json:"id_number" validate:"required"`
	Occupation   string `json:"occupation,omitempty"`
	Relationship string `json:"relationship" validate:"required"`
	Residence    string `json:"residence" validate:"required"`
	Town         string `json:"town,omitempty"`
	WorkPlace    string `json:"work_place,omitempty"`
	WorkPhone    int64  `json:"work_phone,omitempty"`
	ProvidedBy   string `json:"provided_by,omitempty"`
}

func (a *application) updateStudentEmergencyHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	var payload UpdateStudentEmergencyPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	studentEmergencyData := &store.StudentEmergency{
		ID:           payload.ID,
		Firstname:    payload.Firstname,
		Middlename:   payload.Middlename,
		Lastname:     payload.Lastname,
		Phone:        payload.Phone,
		Email:        payload.Email,
		IdNumber:     payload.IdNumber,
		Occupation:   payload.Occupation,
		Relationship: payload.Relationship,
		Residence:    payload.Residence,
		Town:         payload.Town,
		WorkPlace:    payload.WorkPlace,
		WorkPhone:    payload.WorkPhone,
		ProvidedBy:   payload.ProvidedBy,
		StudentID:    student.ID,
		UpdatedAt:    time.Now().String(),
	}

	ctx := r.Context()

	if err := a.store.Students.UpdateStudentEmergency(ctx, *studentEmergencyData, student.ID); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, studentEmergencyData); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

// Guardians
func (a *application) getStudentGuardiansHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	ctx := r.Context()

	studentGuardiansData, err := a.store.Students.GetStudentGuardiansByID(ctx, student.ID)
	if err != nil {
		a.notFoundError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, studentGuardiansData); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type StudentGuardianPayload struct {
	Title          string `json:"title" validate:"required"`
	Firstname      string `json:"firstname" validate:"required"`
	Lastname       string `json:"lastname" validate:"required"`
	Middlename     string `json:"middlename,omitempty"`
	Phone          int64  `json:"phone" validate:"required"`
	PhoneAlternate *int64 `json:"phone_alternate,omitempty"`
	Email          string `json:"email,omitempty"`
	IdNumber       int64  `json:"id_number" validate:"required"`
	KraPinNo       string `json:"kra_pin_no,omitempty"`
	PassportNo     string `json:"passport_no,omitempty"`
	AlienNo        string `json:"alien_no,omitempty"`
	Occupation     string `json:"occupation,omitempty"`
	WorkLocation   string `json:"work_location,omitempty"`
	WorkPhone      *int64 `json:"work_phone,omitempty"`
	Relationship   string `json:"relationship" validate:"required"`
	Address        string `json:"address,omitempty"`
	Residence      string `json:"residence" validate:"required"`
	Town           string `json:"town" validate:"required"`
	County         string `json:"county" validate:"required"`
	SubCounty      string `json:"sub_county" validate:"required"`
	Ward           string `json:"ward,omitempty"`
	VotersCardNo   string `json:"voters_card_no,omitempty"`
	PollingStation string `json:"polling_station,omitempty"`
}

func (a *application) createStudentGuardiansHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	var payload StudentGuardianPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	studentGuardianData := &store.StudentGuardian{
		Title:          payload.Title,
		Firstname:      payload.Firstname,
		Middlename:     payload.Middlename,
		Lastname:       payload.Lastname,
		Phone:          payload.Phone,
		PhoneAlternate: payload.PhoneAlternate,
		Email:          payload.Email,
		IdNumber:       payload.IdNumber,
		KraPinNo:       payload.KraPinNo,
		PassportNo:     payload.PassportNo,
		AlienNo:        payload.AlienNo,
		Occupation:     payload.Occupation,
		WorkLocation:   payload.WorkLocation,
		WorkPhone:      payload.WorkPhone,
		Relationship:   payload.Relationship,
		Address:        payload.Address,
		Residence:      payload.Residence,
		Town:           payload.Town,
		County:         payload.County,
		SubCounty:      payload.SubCounty,
		Ward:           payload.Ward,
		VotersCardNo:   payload.VotersCardNo,
		PollingStation: payload.PollingStation,
		StudentID:      student.ID,
		UpdatedAt:      time.Now().String(),
	}

	ctx := r.Context()

	if err := a.store.Students.CreateStudentGuardian(ctx, *studentGuardianData, student.ID); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, studentGuardianData); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type UpdateStudentGuardianPayload struct {
	ID             int64  `json:"id" validate:"required"`
	Title          string `json:"title" validate:"required"`
	Firstname      string `json:"firstname" validate:"required"`
	Middlename     string `json:"middlename,omitempty"`
	Lastname       string `json:"lastname" validate:"required"`
	Phone          int64  `json:"phone" validate:"required"`
	PhoneAlternate *int64 `json:"phone_alternate,omitempty"`
	Email          string `json:"email,omitempty"`
	IdNumber       int64  `json:"id_number" validate:"required"`
	KraPinNo       string `json:"kra_pin_no,omitempty"`
	PassportNo     string `json:"passport_no,omitempty"`
	AlienNo        string `json:"alien_no,omitempty"`
	Occupation     string `json:"occupation,omitempty"`
	WorkLocation   string `json:"work_location,omitempty"`
	WorkPhone      *int64 `json:"work_phone,omitempty"`
	Relationship   string `json:"relationship" validate:"required"`
	Address        string `json:"address,omitempty"`
	Residence      string `json:"residence" validate:"required"`
	Town           string `json:"town,omitempty"`
	County         string `json:"county,omitempty"`
	SubCounty      string `json:"sub_county,omitempty"`
	Ward           string `json:"ward,omitempty"`
	VotersCardNo   string `json:"voters_card_no,omitempty"`
	PollingStation string `json:"polling_station,omitempty"`
}

func (a *application) updateStudentGuardiansHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	var payload UpdateStudentGuardianPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	studentGuardianData := &store.StudentGuardian{
		ID:             payload.ID,
		Title:          payload.Title,
		Firstname:      payload.Firstname,
		Middlename:     payload.Middlename,
		Lastname:       payload.Lastname,
		Phone:          payload.Phone,
		PhoneAlternate: payload.PhoneAlternate,
		Email:          payload.Email,
		IdNumber:       payload.IdNumber,
		KraPinNo:       payload.KraPinNo,
		PassportNo:     payload.PassportNo,
		AlienNo:        payload.AlienNo,
		Occupation:     payload.Occupation,
		WorkLocation:   payload.WorkLocation,
		WorkPhone:      payload.WorkPhone,
		Relationship:   payload.Relationship,
		Address:        payload.Address,
		Residence:      payload.Residence,
		Town:           payload.Town,
		County:         payload.County,
		SubCounty:      payload.SubCounty,
		Ward:           payload.Ward,
		VotersCardNo:   payload.VotersCardNo,
		PollingStation: payload.PollingStation,
		StudentID:      student.ID,
		UpdatedAt:      time.Now().String(),
	}

	ctx := r.Context()

	if err := a.store.Students.UpdateStudentGuardian(ctx, *studentGuardianData, student.ID); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, studentGuardianData); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type DeleteStudentGuardianPayload struct {
	ID int64 `json:"id" validate:"required"`
}

func (a *application) deleteStudentGuardiansHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	var payload DeleteStudentGuardianPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	if err := a.store.Students.DeleteStudentGuardian(ctx, payload.ID, student.ID); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusNoContent, ""); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

func (a *application) uploadDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	// student := getStudentFromCtx(r)

	ctx := r.Context()

	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		a.badRequestError(w, r, err)
		return
	}

	file, header, err := r.FormFile("declaration")
	if err != nil {
		a.badRequestError(w, r, err)
		return
	}
	file1, header1, err := r.FormFile("death_cert")
	if err != nil {
		a.badRequestError(w, r, err)
		return
	}
	file2, _, err := r.FormFile("birth_cert")
	if err != nil {
		a.badRequestError(w, r, err)
		return
	}

	file3, _, _ := r.FormFile("id_front")

	defer file.Close()
	defer file1.Close()
	fmt.Println(file)
	fmt.Println()
	fmt.Print(file1)
	fmt.Println()
	fmt.Print(file2)
	fmt.Println()
	fmt.Print(file3)

	//
	bucketName := goDotEnvVariable("MINIO_BUCKET")

	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		a.badRequestError(w, r, err)
		return
	}

	objectName := header.Filename
	fileBuffer := []byte(buf.Bytes())
	reader := bytes.NewReader(fileBuffer)
	contentType := header.Header["Content-Type"][0]
	fileSize := header.Size

	info, err := a.minio.PutObject(ctx, bucketName, objectName, reader, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		a.internalServerError(w, r, err)
		return
	}

	var buf1 bytes.Buffer
	_, err = io.Copy(&buf1, file1)
	if err != nil {
		a.badRequestError(w, r, err)
		return
	}

	objectName = header1.Filename
	fileBuffer = []byte(buf1.Bytes())
	reader = bytes.NewReader(fileBuffer)
	contentType = header1.Header["Content-Type"][0]
	fileSize = header1.Size

	info1, err := a.minio.PutObject(ctx, bucketName, objectName, reader, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		a.internalServerError(w, r, err)
		return
	}

	uploadResp := struct {
		Declaration string `json:"declaration"`
		DeathCert   string `json:"death_cert"`
		CreatedAt   string `json:"created_at"`
	}{
		Declaration: info.Key,
		DeathCert:   info1.Key,
		CreatedAt:   time.Now().String(),
	}

	if err := jsonResponse(w, http.StatusCreated, uploadResp); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type CreateStudentApplicationPayload struct {
	BusraryID int64 `json:"bursary_id" validate:"required"`
}

func (a *application) createStudentApplicationHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	var payload CreateStudentApplicationPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	if err := a.store.Students.CreateStudentApplication(ctx, payload.BusraryID, student.ID); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, ""); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type WithdrawStudentApplicationPayload struct {
	BusraryID int64 `json:"bursary_id" validate:"required"`
}

func (a *application) withdrawStudentApplicationHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	var payload WithdrawStudentApplicationPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	if err := a.store.Students.WithdrawStudentApplication(ctx, payload.BusraryID, student.ID); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusAccepted, ""); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

func (a *application) getStudentApplicationsHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)

	ctx := r.Context()

	applications, err := a.store.Students.GetStudentApplications(ctx, student.ID)
	if err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, applications); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

func (a *application) getBursaryByIDHandler(w http.ResponseWriter, r *http.Request) {
	bursaryParam := chi.URLParam(r, "bursaryID")
	bursaryID, err := strconv.Atoi(bursaryParam)
	if err != nil {
		a.badRequestError(w, r, err)
		return
	}
	student := getStudentFromCtx(r)

	ctx := r.Context()

	bursaryData, err := a.store.Bursaries.GetBursaryApplications(ctx, int64(bursaryID), student.ID)
	if err != nil {
		a.notFoundError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, bursaryData); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}
