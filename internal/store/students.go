package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Student struct {
	ID             int64    `json: "id"`
	Firstname      string   `json: "firstname"`
	Middlename     string   `json: "middlename"`
	Lastname       string   `json: "lastname"`
	Email          string   `json: "email"`
	Password       password `json: "-"`
	Blocked        bool     `json: "blocked"`
	FirstTimeLogin bool     `json: "first_time_login"`
	Activated      bool     `json: "activated"`
	CreatedAt      string   `json: "created_at"`
	UpdatedAt      string   `json: "updated_at"`
	Role           Role     `json:"role"`
	Personal       StudentPersonal
	Institution    StudentInstitution
}

type StudentPersonal struct {
	ID               int64  `json:"id"`
	Dob              string `json:"dob"`
	Gender           string `json:"gender"`
	Citizenship      string `json:"citizenship"`
	Religion         string `json:"religion"`
	ParentalStatus   string `json:"parental_status"`
	BirthCertNo      string `json:"birth_cert_no"`
	BirthTown        string `json:"birth_town,omitempty"`
	BirthCounty      string `json:"birth_county"`
	BirthSubCounty   string `json:"birth_sub_county"`
	Ward             string `json:"ward"`
	VotersCardNo     string `json:"voters_card_no,omitempty"`
	Residence        string `json:"residence"`
	IDNumber         *int64 `json:"id_number,omitempty"`
	Phone            int64  `json:"phone"`
	KraPinNo         string `json:"kra_pin_no,omitempty"`
	PassportNo       string `json:"passport_no,omitempty"`
	SpecialNeed      int64  `json:"special_need"`
	SpecialNeedsType string `json:"special_needs_type"`
	StudentID        int64  `json:"student_id"`
}

type StudentInstitution struct {
	ID             int64  `json:"id"`
	InstName       string `json:"inst_name"`
	InstType       string `json:"inst_type"`
	Category       string `json:"category,omitempty"`
	Telephone      int64  `json:"telephone"`
	Email          string `json:"email,omitempty"`
	Address        string `json:"address"`
	InstCounty     string `json:"inst_county"`
	InstSubCounty  string `json:"inst_sub_county"`
	InstWard       string `json:"inst_ward,omitempty"`
	PrincipalName  string `json:"principal_name"`
	YearJoined     int64  `json:"year_joined"`
	CurrClassLevel string `json:"curr_class_level"`
	AdmNo          string `json:"adm_no"`
	StudentID      int64  `json:"student_id"`
	BankName       string `json:"bank_name"`
	BankBranch     string `json:"bank_branch"`
	BankAccName    string `json:"bank_acc_name"`
	BankAccNo      int64  `json:"bank_acc_no"`
}

type StudentSponsor struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	SponsorshipType    string `json:"sponsorship_type"`
	SponsorshipNature  string `json:"sponsorship_nature"`
	Phone              int64  `json:"phone"`
	Email              string `json:"email,omitempty"`
	Address            string `json:"address,omitempty"`
	ContactPersonName  string `json:"contact_person_name,omitempty"`
	ContactPersonPhone int64  `json:"contact_person_phone,omitempty"`
	StudentID          int64  `json:"student_id"`
	UpdatedAt          string `json:"updated_at"`
}

type StudentEmergency struct {
	ID           int64  `json:"id"`
	Firstname    string `json:"firstname"`
	Middlename   string `json:"middlename,omitempty"`
	Lastname     string `json:"lastname"`
	Phone        int64  `json:"phone"`
	Email        string `json:"email,omitempty"`
	IdNumber     int64  `json:"id_number"`
	Occupation   string `json:"occupation,omitempty"`
	Relationship string `json:"relationship"`
	Residence    string `json:"residence"`
	Town         string `json:"town,omitempty"`
	WorkPlace    string `json:"work_place,omitempty"`
	WorkPhone    int64  `json:"work_phone,omitempty"`
	ProvidedBy   string `json:"provided_by,omitempty"`
	StudentID    int64  `json:"student_id"`
	UpdatedAt    string `json:"updated_at"`
}

type StudentGuardian struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	Firstname      string `json:"firstname"`
	Lastname       string `json:"lastname"`
	Middlename     string `json:"middlename,omitempty"`
	Phone          int64  `json:"phone"`
	PhoneAlternate *int64 `json:"phone_alternate,omitempty"`
	Email          string `json:"email,omitempty"`
	IdNumber       int64  `json:"id_number"`
	KraPinNo       string `json:"kra_pin_no,omitempty"`
	PassportNo     string `json:"passport_no,omitempty"`
	AlienNo        string `json:"alien_no,omitempty"`
	Occupation     string `json:"occupation,omitempty"`
	WorkLocation   string `json:"work_location,omitempty"`
	WorkPhone      *int64 `json:"work_phone,omitempty"`
	Relationship   string `json:"relationship"`
	Address        string `json:"address,omitempty"`
	Residence      string `json:"residence"`
	Town           string `json:"town"`
	County         string `json:"county"`
	SubCounty      string `json:"sub_county"`
	Ward           string `json:"ward,omitempty"`
	VotersCardNo   string `json:"voters_card_no,omitempty"`
	PollingStation string `json:"polling_station,omitempty"`
	StudentID      int64  `json:"student_id"`
	UpdatedAt      string `json:"updated_at"`
}

type StudentInvitation struct {
	Token     string    `json:"token"`
	StudentID int64     `json:"student_id"`
	Expiry    time.Time `json:"expiry"`
}

type StudentsStore struct {
	db *sql.DB
}

func (s *StudentsStore) RegisterAndInvite(ctx context.Context, student *Student, token string, exp time.Duration) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.Register(ctx, tx, student); err != nil {
			return err
		}

		if err := s.createStudentInvitation(ctx, tx, token, exp, student.ID); err != nil {
			return err
		}

		return nil
	})

}

func (s *StudentsStore) Register(ctx context.Context, tx *sql.Tx, student *Student) error {
	query := `
		INSERT INTO students(
		firstname, lastname, middlename, email, password, blocked, first_time_login, activated, role_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7 ,$8, $9) RETURNING id, created_at;
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	err := tx.QueryRowContext(
		ctx,
		query,
		student.Firstname,
		student.Lastname,
		student.Middlename,
		student.Email,
		student.Password.hash,
		false,
		true,
		false,
		1,
	).Scan(
		&student.ID,
		&student.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *StudentsStore) createStudentInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, studentID int64) error {
	query := `
		INSERT INTO students_invitations(
		token, student_id, expiry)
		VALUES ($1, $2, $3);
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := tx.ExecContext(
		ctx,
		query,
		token,
		studentID,
		time.Now().Add(exp),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *StudentsStore) RollBackNewStudent(ctx context.Context, studentID int64, token string) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		err := s.deleteStudent(ctx, tx, studentID)
		if err != nil {
			return err
		}

		err = s.deleteStudentInvite(ctx, tx, token, studentID)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *StudentsStore) deleteStudent(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `
		DELETE FROM students WHERE id = $1;
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	res, err := tx.ExecContext(
		ctx,
		query,
		id,
	)

	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("No student found with given ID")
	}

	return nil
}

func (s *StudentsStore) deleteStudentInvite(ctx context.Context, tx *sql.Tx, token string, studentID int64) error {
	query := `
		DELETE from students_invitations where token = $1 AND student_id = $2
	`
	fmt.Println(token)

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := tx.ExecContext(
		ctx,
		query,
		token,
		studentID,
	)

	if err != nil {
		fmt.Println(err)
		fmt.Println("from delete student invite rollback")
		return err
	}

	return nil
}

func (s *StudentsStore) Activate(ctx context.Context, token string) error {

	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		invite, err := s.checkInvite(ctx, tx, token)
		if err != nil {
			return err
		}

		if err := s.updateStudent(ctx, tx, invite.StudentID); err != nil {
			return err
		}

		if err := s.deleteStudentInvite(ctx, tx, token, invite.StudentID); err != nil {
			return err
		}

		return nil
	})
}

func (s *StudentsStore) checkInvite(ctx context.Context, tx *sql.Tx, token string) (*StudentInvitation, error) {
	query := `
		SELECT * from students_invitations where token = $1
	`

	var invite StudentInvitation

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err := tx.QueryRowContext(
		ctx,
		query,
		token,
	).Scan(
		&invite.Token,
		&invite.StudentID,
		&invite.Expiry,
	)

	if time.Now().After(invite.Expiry) {
		return nil, errors.New("token has expired or is invalid")
	}

	if err != nil {
		return nil, err
	}

	return &invite, nil
}

func (s *StudentsStore) updateStudent(ctx context.Context, tx *sql.Tx, studentId int64) error {
	query := `
		UPDATE students SET ACTIVATED = true where id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	res, err := tx.ExecContext(
		ctx,
		query,
		studentId,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("no users updated")
	}

	return nil
}

func (s *StudentsStore) GetOneByEmail(ctx context.Context, email string) (*Student, error) {
	query := `
		SELECT id, email, firstname, password, blocked, created_at, updated_at FROM students WHERE email = $1 
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	student := &Student{}
	if err := s.db.QueryRowContext(ctx, query, email).Scan(
		&student.ID,
		&student.Email,
		&student.Firstname,
		&student.Password.hash,
		&student.Blocked,
		&student.CreatedAt,
		&student.UpdatedAt,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("student not found")
		default:
			return nil, err
		}
	}

	return student, nil
}

func (s *StudentsStore) GetOneByID(ctx context.Context, studentID int64) (*Student, error) {
	query := `
		SELECT id, email, firstname, password, blocked, created_at, updated_at FROM students WHERE id = $1 
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	student := &Student{}
	if err := s.db.QueryRowContext(ctx, query, studentID).Scan(
		&student.ID,
		&student.Email,
		&student.Firstname,
		&student.Password.hash,
		&student.Blocked,
		&student.CreatedAt,
		&student.UpdatedAt,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("student not found")
		default:
			return nil, err
		}
	}

	return student, nil
}

func (s *StudentsStore) GetStudentPersonalByID(ctx context.Context, studentID int64) (*StudentPersonal, error) {
	query := `
	SELECT id, dob, gender, citizenship, religion, parental_status, birth_cert_no, birth_town, birth_county, birth_sub_county, ward, voters_card_no, residence, id_number, phone, kra_pin_no, passport_no, special_need, special_needs_type
	FROM students_personal WHERE student_id = $1;	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	studentPersonal := &StudentPersonal{}

	if err := s.db.QueryRowContext(ctx, query, studentID).Scan(
		&studentPersonal.ID,
		&studentPersonal.Dob,
		&studentPersonal.Gender,
		&studentPersonal.Citizenship,
		&studentPersonal.Religion,
		&studentPersonal.ParentalStatus,
		&studentPersonal.BirthCertNo,
		&studentPersonal.BirthTown,
		&studentPersonal.BirthCounty,
		&studentPersonal.BirthSubCounty,
		&studentPersonal.Ward,
		&studentPersonal.VotersCardNo,
		&studentPersonal.Residence,
		&studentPersonal.IDNumber,
		&studentPersonal.Phone,
		&studentPersonal.KraPinNo,
		&studentPersonal.PassportNo,
		&studentPersonal.SpecialNeed,
		&studentPersonal.SpecialNeedsType,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("student personal details not found")
		default:
			return nil, err
		}
	}

	return studentPersonal, nil
}

func (s *StudentsStore) CreateStudentPersonal(ctx context.Context, payload StudentPersonal, studentID int64) error {
	query := `
		INSERT INTO students_personal(
			dob, gender, citizenship, religion, parental_status, birth_cert_no, birth_town, birth_county, birth_sub_county, ward, voters_card_no, residence, id_number, phone, kra_pin_no, passport_no, special_need, special_needs_type, student_id
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		);

	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		payload.Dob,
		payload.Gender,
		payload.Citizenship,
		payload.Religion,
		payload.ParentalStatus,
		payload.BirthCertNo,
		payload.BirthTown,
		payload.BirthCounty,
		payload.BirthSubCounty,
		payload.Ward,
		payload.VotersCardNo,
		payload.Residence,
		payload.IDNumber,
		payload.Phone,
		payload.KraPinNo,
		0, // payload.PassportNo,
		payload.SpecialNeed,
		payload.SpecialNeedsType,
		studentID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *StudentsStore) UpdateStudentPersonal(ctx context.Context, payload StudentPersonal, studentID int64) error {
	query := `
		UPDATE students_personal
		SET 
			dob = $1,
			gender = $2,
			citizenship = $3,
			religion = $4,
			parental_status = $5,
			birth_cert_no = $6,
			birth_town = $7,
			birth_county = $8,
			birth_sub_county = $9,
			ward = $10,
			voters_card_no = $11,
			residence = $12,
			id_number = $13,
			phone = $14,
			kra_pin_no = $15,
			passport_no = $16,
			special_need = $17,
			special_needs_type = $18
		WHERE id = $19;
	`

	fmt.Println("query")
	fmt.Println(query)
	fmt.Println("query")

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		payload.Dob,
		payload.Gender,
		payload.Citizenship,
		payload.Religion,
		payload.ParentalStatus,
		payload.BirthCertNo,
		payload.BirthTown,
		payload.BirthCounty,
		payload.BirthSubCounty,
		payload.Ward,
		payload.VotersCardNo,
		payload.Residence,
		payload.IDNumber,
		payload.Phone,
		payload.KraPinNo,
		payload.PassportNo,
		payload.SpecialNeed,
		payload.SpecialNeedsType,
		payload.ID,
		// studentID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *StudentsStore) GetStudentInstitutionByID(ctx context.Context, studentID int64) (*StudentInstitution, error) {
	query := `
	SELECT id, inst_name, inst_type, category, telephone, email, address, inst_county, inst_sub_county, inst_ward, principal_name, year_joined, curr_class_level, adm_no, student_id, bank_name, bank_branch, bank_acc_name, bank_acc_no
	FROM students_institution WHERE student_id = $1;	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	studentInstitution := &StudentInstitution{}

	if err := s.db.QueryRowContext(ctx, query, studentID).Scan(
		&studentInstitution.ID,
		&studentInstitution.InstName,
		&studentInstitution.InstType,
		&studentInstitution.Category,
		&studentInstitution.Telephone,
		&studentInstitution.Email,
		&studentInstitution.Address,
		&studentInstitution.InstCounty,
		&studentInstitution.InstSubCounty,
		&studentInstitution.InstWard,
		&studentInstitution.PrincipalName,
		&studentInstitution.YearJoined,
		&studentInstitution.CurrClassLevel,
		&studentInstitution.AdmNo,
		&studentInstitution.StudentID,
		&studentInstitution.BankName,
		&studentInstitution.BankBranch,
		&studentInstitution.BankAccName,
		&studentInstitution.BankAccNo,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("student institution details not found")
		default:
			return nil, err
		}
	}

	return studentInstitution, nil
}

func (s *StudentsStore) CreateStudentInstitution(ctx context.Context, payload StudentInstitution, studentID int64) error {
	query := `
		INSERT INTO public.students_institution(
			inst_name, inst_type, category, telephone, email, address, 
			inst_county, inst_sub_county, inst_ward, principal_name, 
			year_joined, curr_class_level, adm_no, bank_name, 
			bank_branch, bank_acc_name, bank_acc_no, student_id
		)
		VALUES 
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18);
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		payload.InstName,
		payload.InstType,
		payload.Category,
		payload.Telephone,
		payload.Email,
		payload.Address,
		payload.InstCounty,
		payload.InstSubCounty,
		payload.InstWard,
		payload.PrincipalName,
		payload.YearJoined,
		payload.CurrClassLevel,
		payload.AdmNo,
		payload.BankName,
		payload.BankBranch,
		payload.BankAccName,
		payload.BankAccNo,
		studentID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *StudentsStore) UpdateStudentInstitution(ctx context.Context, payload StudentInstitution, studentID int64) error {
	query := `
		UPDATE students_institution
		SET 
			inst_name = $2,
			inst_type = $3,
			category = $4,
			telephone = $5,
			email = $6,
			address = $7,
			inst_county = $8,
			inst_sub_county = $9,
			inst_ward = $10,
			principal_name = $11,
			year_joined = $12,
			curr_class_level = $13,
			adm_no = $14,
			bank_name = $15,
			bank_branch = $16,
			bank_acc_name = $17,
			bank_acc_no = $18
		WHERE id = $1;

	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		payload.ID,
		payload.InstName,
		payload.InstType,
		payload.Category,
		payload.Telephone,
		payload.Email,
		payload.Address,
		payload.InstCounty,
		payload.InstSubCounty,
		payload.InstWard,
		payload.PrincipalName,
		payload.YearJoined,
		payload.CurrClassLevel,
		payload.AdmNo,
		payload.BankName,
		payload.BankBranch,
		payload.BankAccName,
		payload.BankAccNo,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *StudentsStore) GetStudentSponsorByID(ctx context.Context, studentID int64) (*StudentSponsor, error) {
	query := `
		SELECT id,name,sponsorship_type,sponsorship_nature,phone,email,
		address,contact_person_name,contact_person_phone,updated_at 
		FROM students_sponsor WHERE student_id = $1;	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	studentSponsor := &StudentSponsor{}

	if err := s.db.QueryRowContext(ctx, query, studentID).Scan(
		&studentSponsor.ID,
		&studentSponsor.Name,
		&studentSponsor.SponsorshipType,
		&studentSponsor.SponsorshipNature,
		&studentSponsor.Phone,
		&studentSponsor.Email,
		&studentSponsor.Address,
		&studentSponsor.ContactPersonName,
		&studentSponsor.ContactPersonPhone,
		&studentSponsor.UpdatedAt,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("student sponsor details not found")
		default:
			return nil, err
		}
	}

	return studentSponsor, nil
}

func (s *StudentsStore) CreateStudentSponsor(ctx context.Context, payload StudentSponsor, studentID int64) error {
	query := `
		INSERT INTO students_sponsor (
			name, 
			sponsorship_type, 
			sponsorship_nature, 
			phone, 
			email, 
			address, 
			contact_person_name, 
			contact_person_phone, 
			student_id, 
			updated_at
		)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		payload.Name,
		payload.SponsorshipType,
		payload.SponsorshipNature,
		payload.Phone,
		payload.Email,
		payload.Address,
		payload.ContactPersonName,
		payload.ContactPersonPhone,
		studentID,
		time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *StudentsStore) UpdateStudentSponsor(ctx context.Context, payload StudentSponsor, studentID int64) error {
	query := `
		UPDATE students_sponsor
		SET
			name = $1,
			sponsorship_type = $2,
			sponsorship_nature = $3,
			phone = $4,
			email = $5,
			address = $6,
			contact_person_name = $7,
			contact_person_phone = $8,
			updated_at = $10
		WHERE
			id = $11 AND student_id= $9 ;
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		payload.Name,
		payload.SponsorshipType,
		payload.SponsorshipNature,
		payload.Phone,
		payload.Email,
		payload.Address,
		payload.ContactPersonName,
		payload.ContactPersonPhone,
		studentID,
		time.Now(),
		payload.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *StudentsStore) GetStudentEmergencyByID(ctx context.Context, studentID int64) (*StudentEmergency, error) {
	query := `
		SELECT id,firstname,middlename,lastname,phone,
		email,id_number,occupation,relationship,residence,town,work_place,
		work_phone,provided_by,student_id,updated_at FROM students_emergency
		 WHERE student_id = $1;	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	studentEmergency := &StudentEmergency{}

	if err := s.db.QueryRowContext(ctx, query, studentID).Scan(
		&studentEmergency.ID,
		&studentEmergency.Firstname,
		&studentEmergency.Middlename,
		&studentEmergency.Lastname,
		&studentEmergency.Phone,
		&studentEmergency.Email,
		&studentEmergency.IdNumber,
		&studentEmergency.Occupation,
		&studentEmergency.Relationship,
		&studentEmergency.Residence,
		&studentEmergency.Town,
		&studentEmergency.WorkPlace,
		&studentEmergency.WorkPhone,
		&studentEmergency.ProvidedBy,
		&studentEmergency.StudentID,
		&studentEmergency.UpdatedAt,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("student emergency details not found")
		default:
			return nil, err
		}
	}

	return studentEmergency, nil
}

func (s *StudentsStore) CreateStudentEmergency(ctx context.Context, payload StudentEmergency, studentID int64) error {
	query := `
	INSERT INTO 
	students_emergency (
		firstname, 
		middlename, 
		lastname, 
		phone, 
		email, 
		id_number, 
		occupation, 
		relationship, 
		residence, 
		town, 
		work_place, 
		work_phone, 
		provided_by, 
		student_id,
		updated_at
	)
	VALUES
	(
		$1, 
		$2, 
		$3, 
		$4, 
		$5, 
		$6, 
		$7, 
		$8, 
		$9, 
		$10, 
		$11, 
		$12, 
		$13, 
		$14,
		$15
	);
`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		payload.Firstname,
		payload.Middlename,
		payload.Lastname,
		payload.Phone,
		payload.Email,
		payload.IdNumber,
		payload.Occupation,
		payload.Relationship,
		payload.Residence,
		payload.Town,
		payload.WorkPlace,
		payload.WorkPhone,
		payload.ProvidedBy,
		studentID,
		time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *StudentsStore) UpdateStudentEmergency(ctx context.Context, payload StudentEmergency, studentID int64) error {
	query := `
		UPDATE 
		students_emergency 
		SET 
			firstname = $2,
			middlename = $3,
			lastname = $4,
			phone = $5,
			email = $6,
			id_number = $7,
			occupation = $8,
			relationship = $9,
			residence = $10,
			town = $11,
			work_place = $12,
			work_phone = $13,
			provided_by = $14,
			updated_at = $16
		WHERE
			id = $1 AND student_id= $15 ;
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		payload.ID,
		payload.Firstname,
		payload.Middlename,
		payload.Lastname,
		payload.Phone,
		payload.Email,
		payload.IdNumber,
		payload.Occupation,
		payload.Relationship,
		payload.Residence,
		payload.Town,
		payload.WorkPlace,
		payload.WorkPhone,
		payload.ProvidedBy,
		studentID,
		time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *StudentsStore) GetStudentGuardiansByID(ctx context.Context, studentID int64) ([]StudentGuardian, error) {
	query := `
    SELECT 
		id,title,firstname,lastname,middlename,phone,
		phone_alternate,email,id_number,kra_pin_no,passport_no,
		alien_no,occupation,work_location,work_phone,relationship,
		address,residence,town,county,sub_county,ward,
		voters_card_no,polling_station,student_id,updated_at 
	FROM students_guardian
		 WHERE student_id = $1;	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, studentID)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("student guardian details not found")
		default:
			return nil, err
		}
	}
	defer rows.Close()

	guardians := []StudentGuardian{}

	for rows.Next() {
		var gd StudentGuardian
		err := rows.Scan(
			&gd.ID,
			&gd.Title,
			&gd.Firstname,
			&gd.Lastname,
			&gd.Middlename,
			&gd.Phone,
			&gd.PhoneAlternate,
			&gd.Email,
			&gd.IdNumber,
			&gd.KraPinNo,
			&gd.PassportNo,
			&gd.AlienNo,
			&gd.Occupation,
			&gd.WorkLocation,
			&gd.WorkPhone,
			&gd.Relationship,
			&gd.Address,
			&gd.Residence,
			&gd.Town,
			&gd.County,
			&gd.SubCounty,
			&gd.Ward,
			&gd.VotersCardNo,
			&gd.PollingStation,
			&gd.StudentID,
			&gd.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		guardians = append(guardians, gd)
	}
	return guardians, nil
}

func (s *StudentsStore) CreateStudentGuardian(ctx context.Context, payload StudentGuardian, studentID int64) error {
	query := `
	INSERT INTO 
	students_guardian (
		title, 
		firstname, 
		lastname, 
		middlename, 
		phone, 
		phone_alternate, 
		email, 
		id_number, 
		kra_pin_no, 
		passport_no, 
		alien_no, 
		occupation, 
		work_location, 
		work_phone, 
		relationship, 
		address, 
		residence, 
		town, 
		county, 
		sub_county, 
		ward, 
		voters_card_no, 
		polling_station, 
		student_id, 
		updated_at
	)
	VALUES
	(
		$1, 
		$2, 
		$3, 
		$4, 
		$5, 
		$6, 
		$7, 
		$8, 
		$9, 
		$10, 
		$11, 
		$12, 
		$13, 
		$14, 
		$15, 
		$16, 
		$17, 
		$18, 
		$19, 
		$20, 
		$21, 
		$22, 
		$23, 
		$24, 
		$25
	);
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		payload.Title,
		payload.Firstname,
		payload.Lastname,
		payload.Middlename,
		payload.Phone,
		payload.PhoneAlternate,
		payload.Email,
		payload.IdNumber,
		payload.KraPinNo,
		payload.PassportNo,
		payload.AlienNo,
		payload.Occupation,
		payload.WorkLocation,
		payload.WorkPhone,
		payload.Relationship,
		payload.Address,
		payload.Residence,
		payload.Town,
		payload.County,
		payload.SubCounty,
		payload.Ward,
		payload.VotersCardNo,
		01, //payload.PollingStation,
		studentID,
		time.Now(),
	)
	fmt.Println(query)
	if err != nil {
		return err
	}

	return nil
}

func (s *StudentsStore) UpdateStudentGuardian(ctx context.Context, payload StudentGuardian, studentID int64) error {
	query := `
		UPDATE 
			students_guardian 
		SET 
			title = $1,
			firstname = $2,
			lastname = $3,
			middlename = $4,
			phone = $5,
			phone_alternate = $6,
			email = $7,
			id_number = $8,
			kra_pin_no = $9,
			passport_no = $10,
			alien_no = $11,
			occupation = $12,
			work_location = $13,
			work_phone = $14,
			relationship = $15,
			address = $16,
			residence = $17,
			town = $18,
			county = $19,
			sub_county = $20,
			ward = $21,
			voters_card_no = $22,
			polling_station = $23,
			updated_at = $25
		WHERE 
			id = $26 AND student_id= $24 ;
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		payload.Title,
		payload.Firstname,
		payload.Lastname,
		payload.Middlename,
		payload.Phone,
		payload.PhoneAlternate,
		payload.Email,
		payload.IdNumber,
		payload.KraPinNo,
		payload.PassportNo,
		payload.AlienNo,
		payload.Occupation,
		payload.WorkLocation,
		payload.WorkPhone,
		payload.Relationship,
		payload.Address,
		payload.Residence,
		payload.Town,
		payload.County,
		payload.SubCounty,
		payload.Ward,
		payload.VotersCardNo,
		0, //payload.PollingStation,
		studentID,
		time.Now(),
		payload.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *StudentsStore) DeleteStudentGuardian(ctx context.Context, guardianId int64, studentID int64) error {
	query := `
		DELETE FROM 
			students_guardian 
		WHERE 
			id = $1 AND student_id= $2 ;
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		guardianId,
		studentID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *StudentsStore) CreateStudentApplication(ctx context.Context, bursaryID int64, studentID int64) error {
	query := `
	INSERT INTO applications (
		 bursary_id, student_id, stage, created_at, updated_at
	)
	VALUES ($1, $2, $3, $4, $5);
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		bursaryID,
		studentID,
		"submitted",
		time.Now(),
		time.Now(),
	)
	fmt.Println(query)
	if err != nil {
		return err
	}

	return nil
}

func (s *StudentsStore) WithdrawStudentApplication(ctx context.Context, bursaryID int64, studentID int64) error {
	query := `
	UPDATE applications 
		SET soft_delete = true, updated_at = $3
		WHERE bursary_id = $1 AND student_id = $2
		;
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		bursaryID,
		studentID,
		time.Now(),
	)
	fmt.Println(query)
	if err != nil {
		return err
	}

	return nil
}

func (s *StudentsStore) GetStudentApplications(ctx context.Context, studentID int64) ([]BursaryWithMetadata, error) {
	query := `
	SELECT 		 
		applications.id,
		applications.bursary_id, applications.student_id, 
		applications.stage, applications.remarks, 
		applications.soft_delete, applications.created_at, 
		applications.updated_at,
		bursaries.id,
		bursaries.bursary_name,
		bursaries.end_date
 	FROM applications
	JOIN bursaries ON bursaries.id = applications.bursary_id
	WHERE applications.student_id = $1
	;
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	applicationData := []BursaryWithMetadata{}

	rows, err := s.db.QueryContext(
		ctx,
		query,
		studentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		appl := Application{}
		burs := Bursary{}
		data := BursaryWithMetadata{}

		err := rows.Scan(
			&appl.ID,
			&appl.BursaryID,
			&appl.StudentID,
			&appl.Stage,
			&appl.Remarks,
			&appl.SoftDelete,
			&appl.CreatedAt,
			&appl.UpdatedAt,
			&burs.ID,
			&burs.BursaryName,
			&burs.EndDate,
		)
		if err != nil {
			return nil, err
		}

		data.Bursary = burs
		data.Application = appl

		applicationData = append(applicationData, data)
	}

	return applicationData, nil
}
