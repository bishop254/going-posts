package store

import (
	"context"
	"database/sql"
	"time"
)

type Comment struct {
	ID        int64  `json: "id"`
	PostID    int64  `json: "post_id"`
	UserID    int64  `json: "user_id"`
	User      User   `json: "user"`
	Content   string `json: "content"`
	CreatedAt string `json: "created_at"`
}

type CommentStore struct {
	db *sql.DB
}

func (s *CommentStore) GetPostWithCommentsByID(ctx context.Context, postID int64) ([]Comment, error) {
	query := `
		SELECT comments.id, comments.content, comments.created_at, users.username, users.id FROM comments 
		JOIN users on users.id = comments.user_id 
		WHERE comments.post_id = $1 
		ORDER BY comments.created_at DESC;
	`
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	comments := []Comment{}

	for rows.Next() {
		var c Comment
		c.User = User{}

		err := rows.Scan(
			&c.ID,
			&c.Content,
			&c.CreatedAt,
			&c.User.Username,
			&c.User.ID,
		)
		if err != nil {
			return nil, err
		}

		comments = append(comments, c)
	}

	return comments, nil
}
func (s *CommentStore) Create(ctx context.Context, comment *Comment) error {
	query := `
		INSERT INTO comments (content, post_id, user_id)
		VALUES ($1, $2, $3) RETURNING id, content, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		comment.Content,
		comment.PostID,
		comment.UserID,
	).Scan(
		&comment.ID,
		&comment.Content,
		&comment.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}
