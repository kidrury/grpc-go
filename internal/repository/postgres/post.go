package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kidrury/grpc-blog/internal/domain"
)

type PostRepository struct {
	pool *pgxpool.Pool
}

func NewPostRepository(pool *pgxpool.Pool) PostRepository {
	return PostRepository{
		pool: pool,
	}
}

func (r *PostRepository) Create(ctx context.Context, post *domain.Post) error {
	query := `
		INSERT INTO posts (id, title, content, created_at, created_by)
	`

	_, err := r.pool.Exec(ctx, query, post.ID, post.Title, post.CreatedAt, post.CreatedBy)

	return err
}

func (r *PostRepository) GetByID(ctx context.Context, id string) (*domain.Post, error) {
	query := `
		SELECT id, title, content, created_at, created_by
		FROM posts 
		WHERE id = $1
	`

	post := domain.Post{}

	row := r.pool.QueryRow(ctx, query, id)

	err := row.Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.CreatedBy)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPostNotFound
		}
		return nil, err
	}

	return &post, nil
}

func (r *PostRepository) List(ctx context.Context) ([]*domain.Post, error) {
	query := `
		SELECT id, title, content, created_at, created_by
		FROM posts
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	posts := make([]*domain.Post, 0)

	for rows.Next() {
		post := domain.Post{}

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.CreatedBy,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, err
}

func (r *PostRepository) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM posts WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, id)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrPostNotFound
	}

	return nil
}
