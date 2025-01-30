package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type AdminsStore struct {
	db *sql.DB
}

type Admin struct {
	ID             int64    `json: "id"`
	Firstname      string   `json: "firstname"`
	Middlename     *string  `json: "middlename"`
	Lastname       string   `json: "lastname"`
	Email          string   `json: "email"`
	Password       password `json: "-"`
	Blocked        bool     `json: "blocked"`
	FirstTimeLogin bool     `json: "first_time_login"`
	Activated      bool     `json: "activated"`
	CreatedAt      string   `json: "created_at"`
	UpdatedAt      string   `json: "updated_at"`
	Role           Role     `json:"role"`
	RoleCode       *string  `json:"role_code"`
}

type AdminInvitation struct {
	Token   string    `json:"token"`
	AdminID int64     `json:"admin_id"`
	Expiry  time.Time `json:"expiry"`
}

func (s *AdminsStore) RegisterAndInvite(ctx context.Context, admin *Admin, token string, exp time.Duration) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.Register(ctx, tx, admin); err != nil {
			return err
		}

		if err := s.createAdminInvitation(ctx, tx, token, exp, admin.ID); err != nil {
			return err
		}

		return nil
	})

}

func (s *AdminsStore) Register(ctx context.Context, tx *sql.Tx, admin *Admin) error {
	query := `
		INSERT INTO system_users(
		firstname, lastname, middlename, email, password, blocked, first_time_login, activated, role_id, role_code)
		VALUES ($1, $2, $3, $4, $5, $6, $7 ,$8, $9, $10) RETURNING id, created_at;
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	err := tx.QueryRowContext(
		ctx,
		query,
		admin.Firstname,
		admin.Lastname,
		admin.Middlename,
		admin.Email,
		admin.Password.hash,
		false,
		true,
		false,
		admin.Role.ID,
		admin.RoleCode,
	).Scan(
		&admin.ID,
		&admin.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *AdminsStore) createAdminInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, adminID int64) error {
	query := `
		INSERT INTO admins_invitations(
		token, admin_id, expiry)
		VALUES ($1, $2, $3);
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := tx.ExecContext(
		ctx,
		query,
		token,
		adminID,
		time.Now().Add(exp),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *AdminsStore) RollBackNewAdmin(ctx context.Context, adminID int64, token string) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		err := s.deleteAdmin(ctx, tx, adminID)
		if err != nil {
			return err
		}

		err = s.deleteAdminInvite(ctx, tx, token, adminID)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *AdminsStore) deleteAdmin(ctx context.Context, tx *sql.Tx, adminID int64) error {
	query := `
		DELETE FROM system_users WHERE id = $1;
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	res, err := tx.ExecContext(
		ctx,
		query,
		adminID,
	)

	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("No admin found with given ID")
	}

	return nil
}

func (s *AdminsStore) deleteAdminInvite(ctx context.Context, tx *sql.Tx, token string, adminID int64) error {
	query := `
		DELETE from admins_invitations where token = $1 AND admin_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := tx.ExecContext(
		ctx,
		query,
		token,
		adminID,
	)

	if err != nil {
		fmt.Println(err)
		fmt.Println("from delete admin invite rollback")
		return err
	}

	return nil
}

func (s *AdminsStore) Activate(ctx context.Context, token string) error {

	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		invite, err := s.checkInvite(ctx, tx, token)
		if err != nil {
			return err
		}

		if err := s.updateAdmin(ctx, tx, invite.AdminID); err != nil {
			return err
		}

		if err := s.deleteAdminInvite(ctx, tx, token, invite.AdminID); err != nil {
			return err
		}

		return nil
	})
}

func (s *AdminsStore) checkInvite(ctx context.Context, tx *sql.Tx, token string) (*AdminInvitation, error) {
	query := `
		SELECT * from admins_invitations where token = $1
	`

	var invite AdminInvitation

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err := tx.QueryRowContext(
		ctx,
		query,
		token,
	).Scan(
		&invite.Token,
		&invite.AdminID,
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

func (s *AdminsStore) updateAdmin(ctx context.Context, tx *sql.Tx, adminID int64) error {
	query := `
		UPDATE system_users SET ACTIVATED = true where id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	res, err := tx.ExecContext(
		ctx,
		query,
		adminID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("no admins updated")
	}

	return nil
}

func (s *AdminsStore) GetOneByEmail(ctx context.Context, email string) (*Admin, error) {
	query := `
		SELECT id, email, firstname, role_code, password, blocked, created_at, updated_at FROM system_users WHERE email = $1 
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	admin := &Admin{}
	if err := s.db.QueryRowContext(ctx, query, email).Scan(
		&admin.ID,
		&admin.Email,
		&admin.Firstname,
		&admin.RoleCode,
		&admin.Password.hash,
		&admin.Blocked,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("admin not found")
		default:
			return nil, err
		}
	}

	return admin, nil
}

func (s *AdminsStore) GetAdminDataByID(ctx context.Context, adminID int64) (*Admin, error) {
	adminUser := Admin{}

	err := withTx(s.db, ctx, func(tx *sql.Tx) error {
		adminData, err := s.GetOneByID(ctx, tx, adminID)
		if err != nil {
			return err
		}

		roleData, err := s.GetRoleByID(ctx, tx, adminData.Role.ID)
		if err != nil {
			return err
		}

		adminUser = *adminData
		adminUser.Role = *roleData

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &adminUser, err

}

func (s *AdminsStore) GetOneByID(ctx context.Context, tx *sql.Tx, adminID int64) (*Admin, error) {
	query := `
		SELECT system_users.id, system_users.email,
		 system_users.firstname, system_users.password,
		 system_users.blocked,
		  system_users.role_code, 
		  system_users.created_at, system_users.updated_at,
		  roles.id, 
		  roles.name, 
		  roles.level
		FROM system_users 
		JOIN roles ON roles.id = system_users.role_id
		WHERE system_users.id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	admin := &Admin{}
	rl := &Role{}

	if err := s.db.QueryRowContext(ctx, query, adminID).Scan(
		&admin.ID,
		&admin.Email,
		&admin.Firstname,
		&admin.Password.hash,
		&admin.Blocked,
		&admin.RoleCode,
		&admin.CreatedAt,
		&admin.UpdatedAt,
		&rl.ID,
		&rl.Name,
		&rl.Level,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("admin not found")
		default:
			return nil, err
		}
	}

	admin.Role = *rl

	return admin, nil
}

func (s *AdminsStore) GetAdminUsers(ctx context.Context, pq *PaginatedAdminUserQuery, level int64) ([]Admin, error) {
	//TODO : add query parameters
	query := `
		SELECT 
		 system_users.id, system_users.email,
		 system_users.firstname, system_users.middlename,
		 system_users.lastname,
		 system_users.blocked,
		  system_users.role_code, 
		  system_users.created_at, system_users.updated_at,
		  roles.id, 
		  roles.name, 
		  roles.level
		FROM system_users
		JOIN roles ON roles.id = system_users.role_id
		WHERE roles.level <= $1
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, level)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("admins not found")
		default:
			return nil, err
		}
	}
	defer rows.Close()

	adminList := []Admin{}
	for rows.Next() {
		var admin Admin
		var rl Role

		err := rows.Scan(
			&admin.ID,
			&admin.Email,
			&admin.Firstname,
			&admin.Middlename,
			&admin.Lastname,
			&admin.Blocked,
			&admin.RoleCode,
			&admin.CreatedAt,
			&admin.UpdatedAt,
			&rl.ID,
			&rl.Name,
			&rl.Level,
		)
		if err != nil {
			return nil, err
		}

		admin.Role = rl
		adminList = append(adminList, admin)
	}

	return adminList, nil
}

func (s *AdminsStore) GetRoles(ctx context.Context, level int64) ([]Role, error) {
	//TODO : add query parameters
	query := `
		SELECT id, name, description, level 
		FROM roles where level <= $1
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, level)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("roles not found")
		default:
			return nil, err
		}
	}
	defer rows.Close()

	roleList := []Role{}
	for rows.Next() {
		var rl Role

		err := rows.Scan(
			&rl.ID,
			&rl.Name,
			&rl.Description,
			&rl.Level,
		)

		if err != nil {
			return nil, err
		}

		roleList = append(roleList, rl)
	}

	return roleList, nil
}

func (s *AdminsStore) GetRoleByID(ctx context.Context, tx *sql.Tx, roleID int64) (*Role, error) {
	//TODO : add query parameters
	query := `
		SELECT id, name, description, level 
		FROM roles where id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	rl := &Role{}
	err := s.db.QueryRowContext(ctx, query, roleID).Scan(
		&rl.ID,
		&rl.Name,
		&rl.Description,
		&rl.Level,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("role not found")
		default:
			return nil, err
		}
	}

	return rl, nil
}
