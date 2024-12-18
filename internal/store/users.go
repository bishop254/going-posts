package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int64    `json: "id"`
	Username       string   `json: "username"`
	Email          string   `json: "email"`
	Password       password `json: "-"`
	Blocked        bool     `json: "blocked"`
	FirstTimeLogin bool     `json: "first_time_login"`
	CreatedAt      string   `json: "created_at"`
	UpdatedAt      string   `json: "updated_at"`
}

type Invitation struct {
	Token  string    `json:"token"`
	UserID int64     `json:"user_id"`
	Expiry time.Time `json:"expiry"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Hashing(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		INSERT INTO users (username, email, password) VALUES($1, $2, $3) RETURNING id, username, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password.hash,
	).Scan(
		&user.ID,
		&user.Username,
		&user.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}
func (s *UserStore) GetOne(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, email, username, blocked, created_at, updated_at FROM users WHERE id = $1 
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	var user User

	if err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Blocked,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("user not found")
		default:
			return nil, err
		}
	}

	return &user, nil
}
func (s *UserStore) FollowUser(ctx context.Context, userId int64, followerId int64) error {
	query := `
		INSERT INTO followers(user_id, follower_id) VALUES($1, $2) RETURNING created_at
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	res, err := s.db.ExecContext(
		ctx,
		query,
		userId,
		followerId,
	)
	if err != nil {
		fmt.Print(err, "from repository function")
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("user/follower not found")
	}

	return nil
}
func (s *UserStore) UnfollowUser(ctx context.Context, userId int64, followerId int64) error {
	query := `
		DELETE FROM followers
		WHERE user_id = $1 AND follower_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	res, err := s.db.ExecContext(
		ctx,
		query,
		userId,
		followerId,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if rows == 0 {
		return errors.New("resource not found")
	}

	return nil
}

func (s *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, expiry time.Duration) error {

	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}

		if err := s.createUserInvitation(ctx, tx, token, expiry, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userId int64) error {
	query := `
		INSERT INTO invitations(token, user_id, expiry)
		VALUES ($1, $2, $3)
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := tx.ExecContext(
		ctx,
		query,
		token,
		userId,
		time.Now().Add(exp),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) Activate(ctx context.Context, token string) error {

	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		invite, err := s.checkInvite(ctx, tx, token)
		if err != nil {
			return err
		}

		if err := s.updateUser(ctx, tx, invite.UserID); err != nil {
			return err
		}

		if err := s.deleteInvite(ctx, tx, token, invite.UserID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) checkInvite(ctx context.Context, tx *sql.Tx, token string) (*Invitation, error) {
	query := `
		SELECT * from invitations where token = $1
	`

	var invite Invitation

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err := tx.QueryRowContext(
		ctx,
		query,
		token,
	).Scan(
		&invite.Token,
		&invite.UserID,
		&invite.Expiry,
	)

	if time.Now().After(invite.Expiry) {
		return nil, errors.New("token has expired")
	}

	if err != nil {
		return nil, err
	}

	return &invite, nil
}

func (s *UserStore) updateUser(ctx context.Context, tx *sql.Tx, userId int64) error {
	query := `
		UPDATE users SET FIRST_TIME_LOGIN = true where id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	res, err := tx.ExecContext(
		ctx,
		query,
		userId,
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

func (s *UserStore) deleteInvite(ctx context.Context, tx *sql.Tx, token string, userId int64) error {
	query := `
		DELETE from invitations where token = $1 AND user_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := tx.ExecContext(
		ctx,
		query,
		token,
		userId,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) RollBackNewUser(ctx context.Context, id int64, token string) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		err := s.deleteUser(ctx, tx, id)
		if err != nil {
			return err
		}

		err = s.deleteInvite(ctx, tx, token, id)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) deleteUser(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `
		DELETE FROM users WHERE id = $1;
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
		return errors.New("No user found with given ID")
	}

	return nil
}
