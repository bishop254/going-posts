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
	SpecialNeed      bool   `json:"special_need"`
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
