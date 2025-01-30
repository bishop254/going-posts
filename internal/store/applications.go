package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type ApplicationsStore struct {
	db *sql.DB
}

func (s *ApplicationsStore) GetApplications(ctx context.Context, tableName string, stage string) ([]ApplicationWithMetadata, error) {
	allowedTables := map[string]bool{
		"applications":             true,
		"county_applications":      true,
		"min_finance_applications": true,
	}

	if !allowedTables[tableName] {
		return nil, errors.New("invalid table name")
	}

	query := fmt.Sprintf(`
		SELECT 		 
			app.id,
			app.bursary_id,
			app.student_id, 
			app.stage,
			app.remarks, 
			app.soft_delete,
			app.created_at, 
			app.updated_at,
			bursaries.id AS bursary_id,
			bursaries.bursary_name,
			bursaries.end_date,
			students.firstname,
			students.lastname,
			students_personal.gender,
			students_personal.phone,
			students_personal.birth_sub_county,
			students_personal.ward,
			students_institution.inst_name,
			students_institution.adm_no
		FROM %s AS app
		LEFT JOIN students ON students.id = app.student_id
		LEFT JOIN students_personal ON students_personal.student_id = app.student_id
		LEFT JOIN students_institution ON students_institution.student_id = app.student_id
		JOIN bursaries ON bursaries.id = app.bursary_id
		WHERE app.soft_delete = false AND app.stage = $1 ;
	`, tableName)

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	applicationData := []ApplicationWithMetadata{}

	rows, err := s.db.QueryContext(
		ctx,
		query,
		stage,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		appl := Application{}
		burs := Bursary{}
		stud := Student{}
		studPers := StudentPersonal{}
		studInst := StudentInstitution{}

		data := ApplicationWithMetadata{}

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
			&stud.Firstname,
			&stud.Lastname,
			&studPers.Gender,
			&studPers.Phone,
			&studPers.BirthSubCounty,
			&studPers.Ward,
			&studInst.InstName,
			&studInst.AdmNo,
		)
		if err != nil {
			return nil, err
		}

		stud.Personal = studPers
		stud.Institution = studInst

		data.Bursary = burs
		data.Application = appl
		data.Student = stud

		applicationData = append(applicationData, data)
	}

	return applicationData, nil
}

func (s *ApplicationsStore) GetAllApplications(ctx context.Context) ([]ApplicationWithMetadata, error) {
	query := fmt.Sprintf(`
		SELECT 		 
			app.id,
			app.bursary_id,
			app.student_id, 
			app.stage,
			app.remarks, 
			app.soft_delete,
			app.created_at, 
			app.updated_at,
			bursaries.id AS bursary_id,
			bursaries.bursary_name,
			bursaries.end_date,
			students.firstname,
			students.lastname,
			students_personal.gender,
			students_personal.phone,
			students_personal.birth_sub_county,
			students_personal.ward,
			students_institution.inst_name,
			students_institution.adm_no
		FROM %s AS app
		LEFT JOIN students ON students.id = app.student_id
		LEFT JOIN students_personal ON students_personal.student_id = app.student_id
		LEFT JOIN students_institution ON students_institution.student_id = app.student_id
		JOIN bursaries ON bursaries.id = app.bursary_id
		WHERE app.soft_delete = false ;
	`, "applications")

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	applicationData := []ApplicationWithMetadata{}

	rows, err := s.db.QueryContext(
		ctx,
		query,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		appl := Application{}
		burs := Bursary{}
		stud := Student{}
		studPers := StudentPersonal{}
		studInst := StudentInstitution{}

		data := ApplicationWithMetadata{}

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
			&stud.Firstname,
			&stud.Lastname,
			&studPers.Gender,
			&studPers.Phone,
			&studPers.BirthSubCounty,
			&studPers.Ward,
			&studInst.InstName,
			&studInst.AdmNo,
		)
		if err != nil {
			return nil, err
		}

		stud.Personal = studPers
		stud.Institution = studInst

		data.Bursary = burs
		data.Application = appl
		data.Student = stud

		applicationData = append(applicationData, data)
	}

	return applicationData, nil
}

func (s *ApplicationsStore) GetApplicationMetaDataByID(ctx context.Context, tableName string, applicationID int64) (*ApplicationWithMetadata, error) {
	allowedTables := map[string]bool{
		"applications":             true,
		"county_applications":      true,
		"min_finance_applications": true,
	}

	if !allowedTables[tableName] {
		return nil, errors.New("invalid table name")
	}

	query := fmt.Sprintf(`
		SELECT 		 
			app.id,
			app.bursary_id,
			app.student_id, 
			app.stage,
			app.remarks, 
			app.soft_delete,
			app.created_at, 
			app.updated_at,
			bursaries.id AS bursary_id,
			bursaries.bursary_name,
			bursaries.description, 
			bursaries.end_date,
			bursaries.amount_allocated, 
			bursaries.amount_per_student, 
			bursaries.allocation_type, 
			bursaries.created_at AS bursary_created_at,
			students.id,
			students.firstname,
			students.middlename,
			students.lastname,
			students.email,
			students_personal.dob,
			students_personal.gender,
			students_personal.parental_status,
			students_personal.birth_sub_county,
			students_personal.ward,
			students_personal.residence,
			students_personal.phone,
			students_personal.id_number,
			students_personal.special_need,
			students_institution.inst_name,
			students_institution.inst_type,
			students_institution.category,
			students_institution.inst_county,
			students_institution.inst_sub_county,
			students_institution.inst_ward,
			students_institution.adm_no,
			students_institution.bank_name,
			students_institution.bank_branch,
			students_institution.bank_acc_name,
			students_institution.bank_acc_no
		FROM %s AS app
		LEFT JOIN students ON students.id = app.student_id
		LEFT JOIN students_personal ON students_personal.student_id = app.student_id
		LEFT JOIN students_institution ON students_institution.student_id = app.student_id
		LEFT JOIN bursaries ON bursaries.id = app.bursary_id
		WHERE app.soft_delete = false AND app.id = $1;
	`, tableName)

	// query := `
	// SELECT
	// 	applications.id,
	// 	applications.bursary_id,
	// 	applications.student_id,
	// 	applications.stage,
	// 	applications.remarks,
	// 	applications.soft_delete,
	// 	applications.created_at,
	// 	applications.updated_at,
	// 	bursaries.id,
	// 	bursaries.bursary_name,
	// 	bursaries.description,
	// 	bursaries.end_date,
	// 	bursaries.amount_allocated,
	// 	bursaries.amount_per_student,
	// 	bursaries.allocation_type,
	// 	bursaries.created_at,
	//     students.firstname,
	//     students.middlename,
	//     students.lastname,
	//     students.email,
	//     students_personal.dob,
	//     students_personal.gender,
	//     students_personal.parental_status,
	//     students_personal.birth_sub_county,
	//     students_personal.ward,
	//     students_personal.residence,
	//     students_personal.phone,
	//     students_personal.id_number,
	//     students_personal.special_need,
	//     students_institution.inst_name,
	//     students_institution.inst_type,
	//     students_institution.category,
	//     students_institution.inst_county,
	//     students_institution.inst_sub_county,
	//     students_institution.inst_ward,
	//     students_institution.adm_no,
	//     students_institution.bank_name,
	//     students_institution.bank_branch,
	//     students_institution.bank_acc_name,
	//     students_institution.bank_acc_no
	// FROM applications
	// LEFT JOIN students ON students.id = applications.student_id
	// LEFT JOIN students_personal ON students_personal.student_id = applications.student_id
	// LEFT JOIN students_institution ON students_institution.student_id = applications.student_id
	// LEFT JOIN bursaries ON bursaries.id = applications.bursary_id
	// WHERE applications.soft_delete = false AND applications.id = $1;
	// `

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	applicationData := &ApplicationWithMetadata{}
	appl := Application{}
	burs := Bursary{}
	stud := Student{}
	studPers := StudentPersonal{}
	studInst := StudentInstitution{}

	if err := s.db.QueryRowContext(
		ctx,
		query,
		applicationID,
	).Scan(
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
		&burs.Description,
		&burs.EndDate,
		&burs.AmountAllocated,
		&burs.AmountPerStudent,
		&burs.AllocationType,
		&burs.CreatedAt,
		&stud.ID,
		&stud.Firstname,
		&stud.Middlename,
		&stud.Lastname,
		&stud.Email,
		&studPers.Dob,
		&studPers.Gender,
		&studPers.ParentalStatus,
		&studPers.BirthSubCounty,
		&studPers.Ward,
		&studPers.Residence,
		&studPers.Phone,
		&studPers.IDNumber,
		&studPers.SpecialNeed,
		&studInst.InstName,
		&studInst.InstType,
		&studInst.Category,
		&studInst.InstCounty,
		&studInst.InstSubCounty,
		&studInst.InstWard,
		&studInst.AdmNo,
		&studInst.BankName,
		&studInst.BankBranch,
		&studInst.BankAccName,
		&studInst.BankAccNo,
	); err != nil {
		return nil, err
	}

	stud.Personal = studPers
	stud.Institution = studInst

	applicationData.Bursary = burs
	applicationData.Application = appl
	applicationData.Student = stud

	return applicationData, nil
}

func (s *ApplicationsStore) ApproveApplicationByID(ctx context.Context, stage string, applicationID int64) error {

	query := `
	UPDATE applications
	SET stage = $1
	WHERE id = $2;
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	if _, err := s.db.ExecContext(
		ctx,
		query,
		stage,
		applicationID,
	); err != nil {
		return err
	}

	return nil
}

func (s *ApplicationsStore) ApproveBulkApplications(ctx context.Context, stage string, applicationIDs []int64) error {

	query := `
	UPDATE applications
	SET stage = $1
	WHERE id = ANY($2);
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	if _, err := s.db.ExecContext(
		ctx,
		query,
		stage,
		pq.Array(applicationIDs),
	); err != nil {
		return err
	}

	return nil
}
