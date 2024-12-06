package store

import (
	"context"
	"database/sql"
	"time"
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetOne(context.Context, int64) (*Post, error)
		Delete(context.Context, int64) error
		Update(context.Context, *Post) error
		GetUserFeed(context.Context, int64, *PaginatedFeedQuery) ([]PostWithMetadata, error)
	}
	Users interface {
		Create(context.Context, *sql.Tx, *User) error
		GetOne(context.Context, int64) (*User, error)
		FollowUser(context.Context, int64, int64) error
		UnfollowUser(context.Context, int64, int64) error
		CreateAndInvite(context.Context, *User, string, time.Duration) error
		Activate(context.Context, string) error
	}
	Comments interface {
		Create(context.Context, *Comment) error
		GetPostWithCommentsByID(context.Context, int64) ([]Comment, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:    &PostStore{db},
		Users:    &UserStore{db},
		Comments: &CommentStore{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
