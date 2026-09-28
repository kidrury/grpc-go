package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kidrury/grpc-blog/internal/domain"
)

type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) error
	GetByID(ctx context.Context, id string) (*domain.Post, error)
	List(ctx context.Context) ([]*domain.Post, error)
	Delete(ctx context.Context, id string) error
}

type PostService struct {
	postRepo PostRepository
}

func NewPostService(pr PostRepository) *PostService {
	return &PostService{
		postRepo: pr,
	}
}

func (s *PostService) CreatePost(ctx context.Context, title, content string) (*domain.Post, error) {
	post := &domain.Post{
		ID:        uuid.NewString(),
		Title:     title,
		Content:   content,
		CreatedAt: time.Now(),
		CreatedBy: "later",
	}
	if err := post.Validate(); err != nil {
		return &domain.Post{}, err
	}

	err := s.postRepo.Create(ctx, post)
	if err != nil {
		return &domain.Post{}, fmt.Errorf("repository.Create: %w\n", err)
	}

	return post, nil
}

func (s *PostService) GetPostByID(ctx context.Context, id string) (*domain.Post, error) {
	if id == "" {
		return nil, domain.ErrIDRequired
	}

	post, err := s.postRepo.GetByID(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("repository.GetById: %w", err)
	}

	return post, nil
}

func (s *PostService) ListPosts(ctx context.Context) ([]*domain.Post, error) {
	posts, err := s.postRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository.List: %w", err)
	}

	return posts, nil
}

func (s *PostService) DeletePost(ctx context.Context, id string) error {
	if id == "" {
		return domain.ErrIDRequired
	}

	if err := s.postRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("repository.Delete: %w", err)
	}

	return nil
}
