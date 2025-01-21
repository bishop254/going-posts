package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type BursariesStore struct {
	db *sql.DB
}

type Application struct {
	ID         int64   `json:"id"`
	BursaryID  int64   `json:"bursary_id"`
	StudentID  int64   `json:"student_id"`
	Stage      string  `json:"stage"`
	Remarks    *string `json:"remarks,omitempty"`
	SoftDelete *bool   `json:"soft_delete,omitempty"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

type Bursary struct {
	ID               int64    `json:"id"`
	BursaryName      string   `json:"bursary_name"`
	Description      *string  `json:"description,omitempty"`
	EndDate          string   `json:"end_date"`
	AmountAllocated  *float64 `json:"amount_allocated,omitempty"`
	AmountPerStudent *float64 `json:"amount_per_student,omitempty"`
	AllocationType   string   `json:"allocation_type"`
	CreatedAt        string   `json:"created_at"`
}

type BursariesWithMetadata struct {
	Bursaries  []Bursary `json:"bursaries"`
	TotalItems int       `json:"total_items"`
}

type BursaryWithMetadata struct {
	Bursary     Bursary     `json:"bursary"`
	Application Application `json:"application"`
}

type ApplicationWithMetadata struct {
	Bursary     Bursary     `json:"bursary"`
	Application Application `json:"application"`
	Student     Student     `json:"student"`
}

// TODO : add bursary count based on query, also fix query
func (s *BursariesStore) GetBursariesAndCount(ctx context.Context, fq *PaginatedFeedQuery) (*BursariesWithMetadata, error) {
	bursaryList := BursariesWithMetadata{}

	err := withTx(s.db, ctx, func(tx *sql.Tx) error {
		totalItems, err := s.GetBursariesCount(ctx, tx, fq)
		if err != nil {
			return err
		}

		bursaries, err := s.GetBursaries(ctx, tx, fq)
		if err != nil {
			return err
		}

		bursaryList.Bursaries = bursaries
		bursaryList.TotalItems = int(totalItems)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &bursaryList, err
}

func (s *BursariesStore) GetBursaries(ctx context.Context, tx *sql.Tx, fq *PaginatedFeedQuery) ([]Bursary, error) {
	var searchCondition string
	var allocationCondition string

	if fq.Search != "" {
		searchCondition = "(bursaries.bursary_name ILIKE '%' || $1 || '%' OR bursaries.description ILIKE '%' || $1 || '%')"
	} else {
		searchCondition = "(bursaries.bursary_name ILIKE '%' AND bursaries.description ILIKE '%') OR (bursaries.bursary_name ILIKE '%' || $1 || '%' OR bursaries.description ILIKE '%' || $1 || '%')  "

	}
	if fq.AllocationType != "" {
		allocationCondition = "(bursaries.allocation_type = $6)"
	} else {
		allocationCondition = "(bursaries.allocation_type ILIKE '%') OR (bursaries.allocation_type = $6)"
	}

	query := `
		SELECT 
			bursaries.id, 
			bursaries.bursary_name, 
			bursaries.description, 
			bursaries.end_date, 
			bursaries.amount_allocated, 
			bursaries.amount_per_student, 
			bursaries.allocation_type, 
			bursaries.created_at
		FROM 
			bursaries
		WHERE 
		` + searchCondition + `
		 AND (bursaries.created_at BETWEEN $2 AND $3) AND 
		` + allocationCondition + ` 
		ORDER BY 
			bursaries.created_at  ` + fq.Sort + `
		LIMIT $4
		OFFSET $5;
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	if fq.Search != "" {

	}
	rows, err := tx.QueryContext(
		ctx,
		query,
		fq.Search,
		fq.Since,
		fq.Until,
		fq.Limit,
		fq.Offset,
		fq.AllocationType,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bursaries := []Bursary{}

	for rows.Next() {
		var br Bursary

		err := rows.Scan(
			&br.ID,
			&br.BursaryName,
			&br.Description,
			&br.EndDate,
			&br.AmountAllocated,
			&br.AmountPerStudent,
			&br.AllocationType,
			&br.CreatedAt,
		)

		if err != nil {
			return nil, err
		}
		bursaries = append(bursaries, br)
	}

	return bursaries, nil
}

func (s *BursariesStore) GetBursariesCount(ctx context.Context, tx *sql.Tx, fq *PaginatedFeedQuery) (int64, error) {
	var searchCondition string
	var allocationCondition string
	var totalItems int64

	if fq.Search != "" {
		searchCondition = "(bursaries.bursary_name ILIKE '%' || $1 || '%' OR bursaries.description ILIKE '%' || $1 || '%')"
	} else {
		searchCondition = "(bursaries.bursary_name ILIKE '%' AND bursaries.description ILIKE '%') OR (bursaries.bursary_name ILIKE '%' || $1 || '%' OR bursaries.description ILIKE '%' || $1 || '%')  "

	}
	if fq.AllocationType != "" {
		allocationCondition = "(bursaries.allocation_type = $4)"
	} else {
		allocationCondition = "(bursaries.allocation_type ILIKE '%') OR (bursaries.allocation_type = $4)"
	}

	query := `
		SELECT 
		    COUNT(*) AS total_items
		FROM 
			bursaries
		WHERE 
		` + searchCondition + `
		 AND (bursaries.created_at BETWEEN $2 AND $3) AND 
		` + allocationCondition + ` ;
	`

	fmt.Println(query)
	fmt.Println(fq.AllocationType)

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	if err := tx.QueryRowContext(
		ctx,
		query,
		fq.Search,
		fq.Since,
		fq.Until,
		fq.AllocationType,
	).Scan(&totalItems); err != nil {
		return 0, err
	}

	return totalItems, nil
}

func (s *BursariesStore) CreateBursary(ctx context.Context, payload Bursary) error {
	query := `
	INSERT INTO bursaries(
		bursary_name, description, end_date, 
		amount_allocated, amount_per_student, 
		allocation_type, created_at
		) 
	VALUES 
	 ($1, $2, $3, $4, $5, $6, $7);
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		payload.BursaryName,
		payload.Description,
		payload.EndDate,
		payload.AmountAllocated,
		payload.AmountPerStudent,
		payload.AllocationType,
		time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *BursariesStore) UpdateBursary(ctx context.Context, payload Bursary) error {
	query := `
	UPDATE bursaries
		SET  
			bursary_name=$1, description=$2, 
			end_date=$3, amount_allocated=$4, 
			amount_per_student=$5, allocation_type=$6		
		WHERE id = $7;
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		payload.BursaryName,
		payload.Description,
		payload.EndDate,
		payload.AmountAllocated,
		payload.AmountPerStudent,
		payload.AllocationType,
		payload.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

// func (s *BursariesStore) GetBursaryAndApplications(ctx context.Context, bursaryID int64) (*BursaryWithMetadata, error) {
// 	bursaryData := BursaryWithMetadata{}

// 	err := withTx(s.db, ctx, func(tx *sql.Tx) error {
// 		totalItems, err := s.GetBursariesCount(ctx, tx, fq)
// 		if err != nil {
// 			return err
// 		}

// 		bursaries, err := s.GetBursaries(ctx, tx, fq)
// 		if err != nil {
// 			return err
// 		}

// 		bursaryList.Bursaries = bursaries
// 		bursaryList.TotalItems = int(totalItems)

// 		return nil
// 	})

// 	if err != nil {
// 		return nil, err
// 	}

// 	return &bursaryList, err
// }

// func (s *BursariesStore) GetBursaryByID(ctx context.Context, tx *sql.Tx, bursaryID int64) (*Bursary, error) {
// 	query := `
// 		SELECT
// 			id,
// 			bursary_name,
// 			description,
// 			end_date,
// 			amount_allocated,
// 			amount_per_student,
// 			allocation_type,
// 			created_at
// 		FROM
// 			bursaries
// 		WHERE
// 		 id = $1
// 	`

// 	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
// 	defer cancel()

// 	bursary := Bursary{}

// 	err := tx.QueryRowContext(
// 		ctx,
// 		query,
// 		bursaryID,
// 	).Scan(
// 		&bursary.ID,
// 		&bursary.BursaryName,
// 		&bursary.Description,
// 		&bursary.EndDate,
// 		&bursary.AmountAllocated,
// 		&bursary.AmountPerStudent,
// 		&bursary.AllocationType,
// 		&bursary.CreatedAt,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &bursary, nil
// }

func (s *BursariesStore) GetBursaryApplications(ctx context.Context, bursaryID int64, studentID int64) ([]BursaryWithMetadata, error) {
	query := `
	SELECT 
			bursaries.id,
			bursaries.bursary_name, 
			bursaries.description, 
			bursaries.end_date, 
			bursaries.amount_allocated, 
			bursaries.amount_per_student, 
			bursaries.allocation_type, 
			bursaries.created_at,
			applications.id,
			applications.bursary_id,
			applications.student_id,
			applications.remarks,
			applications.stage,
			applications.soft_delete,
			applications.created_at,
			applications.updated_at
		FROM 
			bursaries
		JOIN
			applications ON applications.bursary_id = bursaries.id
		WHERE
			bursaries.id = $1 AND applications.student_id = $2
		ORDER BY 
			applications.created_at desc
		
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	rows, err := s.db.QueryContext(
		ctx,
		query,
		bursaryID,
		studentID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	bursaryData := []BursaryWithMetadata{}

	for rows.Next() {
		var bd BursaryWithMetadata
		var br Bursary
		var appl Application

		err := rows.Scan(
			&br.ID,
			&br.BursaryName,
			&br.Description,
			&br.EndDate,
			&br.AmountAllocated,
			&br.AmountPerStudent,
			&br.AllocationType,
			&br.CreatedAt,
			&appl.ID,
			&appl.BursaryID,
			&appl.StudentID,
			&appl.Remarks,
			&appl.Stage,
			&appl.SoftDelete,
			&appl.CreatedAt,
			&appl.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		bd.Bursary = br
		bd.Application = appl

		bursaryData = append(bursaryData, bd)
	}

	if len(bursaryData) <= 0 {
		return nil, errors.New("no bursary applications")
	}

	return bursaryData, nil
}
