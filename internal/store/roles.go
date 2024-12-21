package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"Description"`
	Level       int    `json:"level"`
}

type RolesStore struct {
	db *sql.DB
}

func (s *RolesStore) GetOneByName(ctx context.Context, requiredRole string) (*Role, error) {
	query := `
		SELECT id, name,level FROM roles where name = $1
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	role := &Role{}
	if err := s.db.QueryRowContext(
		ctx,
		query,
		strings.ToLower(requiredRole),
	).Scan(
		&role.ID,
		&role.Name,
		&role.Level,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("role not found")
		default:
			return nil, err
		}
	}

	return role, nil
}

func (s *RolesStore) GetAllAboveLevel(ctx context.Context, requiredRoleLevel int64) ([]Role, error) {
	query := `
		SELECT id, name,level FROM roles WHERE level > $1
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	rows, err := s.db.QueryContext(
		ctx,
		query,
		requiredRoleLevel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rolesList := []Role{}

	for rows.Next() {
		var rl Role

		err := rows.Scan(
			&rl.ID,
			&rl.Name,
			&rl.Level,
		)

		if err != nil {
			return nil, err
		}
		rolesList = append(rolesList, rl)
	}

	return rolesList, nil
}
