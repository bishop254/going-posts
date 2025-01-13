package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	Tags      []string  `json:"tags"`
	UserID    int64     `json:"user_id"`
	Comments  []Comment `json:"comments"`
	Version   int       `json:"version"`
	User      User      `json:"user"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type PostWithMetadata struct {
	Post
	CommentCount int `json:"comment_count"`
}

type PostStore struct {
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO posts (content, title, tags, user_id)
		VALUES ($1, $2, $3, $4) RETURNING id, title, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		pq.Array(post.Tags),
		post.UserID,
	).Scan(
		&post.ID,
		&post.Title,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostStore) GetOne(ctx context.Context, id int64) (*Post, error) {
	query := `
		SELECT id, user_id, title, content, tags, version, created_at, updated_at FROM posts WHERE id = $1 
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	var post Post

	if err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		pq.Array(&post.Tags),
		&post.Version,
		&post.CreatedAt,
		&post.UpdatedAt,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("post not found")
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (s *PostStore) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM posts WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("post not found")
	}

	return nil
}

func (s *PostStore) Update(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts
		SET title = $1, content = $2, version = version + 1
		WHERE id = $3 AND version = $4
		RETURNING version
	`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.ID,
		post.Version,
	).Scan(&post.Version)

	if err != nil {

		switch {
		case errors.Is(err, errors.New("post not found")):
			return errors.New("post not found")
		default:
			return err
		}
	}

	return nil
}

func (s *PostStore) GetUserFeed(ctx context.Context, id int64, fq *PaginatedFeedQuery) ([]PostWithMetadata, error) {
	var tagsCondition string
	// if len(fq.Tags) > 0 {
	// 	tagsCondition = "(posts.tags ILIKE ANY (ARRAY["
	// 	for i, tag := range fq.Tags {
	// 		tagsCondition += "'" + "%" + tag + "%" + "'"
	// 		if i < len(fq.Tags)-1 {
	// 			tagsCondition += ", "
	// 		}
	// 	}
	// 	tagsCondition += "]))"
	// } else {
	tagsCondition = "(posts.tags ILIKE ANY (ARRAY['%%']))"
	// }

	query := `
	 	SELECT 
			posts.id, 
			posts.user_id, 
			posts.title,
			posts.content, 
			posts.created_at,
			posts.version, 
			posts.tags, 
			COUNT(comments.id) AS comments_count
		FROM 
			posts
		LEFT JOIN 
			comments ON comments.post_id = posts.id
		JOIN 
			followers ON followers.follower_id = posts.user_id OR posts.user_id = $1
		WHERE 
			followers.user_id = $1 AND
			(posts.title ILIKE '%' || $2 || '%' OR posts.content ILIKE '%' || $2 || '%') AND
			` + tagsCondition + ` AND
			(posts.created_at BETWEEN $5 AND $6)

		GROUP BY 
			posts.id, posts.user_id, posts.title, posts.content, posts.created_at, posts.version, posts.tags
		ORDER BY 
			posts.created_at ` + fq.Sort + `
		LIMIT $3 
		OFFSET $4;
	 `

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	rows, err := s.db.QueryContext(
		ctx,
		query,
		int64(1),
		fq.Search,
		fq.Limit,
		fq.Offset,
		fq.Since,
		fq.Until,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	feed := []PostWithMetadata{}

	for rows.Next() {
		var fd PostWithMetadata
		fd.Post = Post{}

		err := rows.Scan(
			&fd.ID,
			&fd.UserID,
			&fd.Title,
			&fd.Content,
			&fd.CreatedAt,
			&fd.Version,
			pq.Array(&fd.Tags),
			&fd.CommentCount,
		)
		if err != nil {
			return nil, err
		}
		feed = append(feed, fd)
	}

	return feed, nil
}
