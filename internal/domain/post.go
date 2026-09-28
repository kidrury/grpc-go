package domain

import (
	"errors"
	"time"
)

var (
	ErrPostNotFound    = errors.New("post not found")
	ErrIDRequired      = errors.New("id required")
	ErrTitleRequired   = errors.New("title required")
	ErrContentRequired = errors.New("content required")
)

type Post struct {
	ID        string
	Title     string
	Content   string
	CreatedBy string
	CreatedAt time.Time
}

func (p *Post) Validate() error {
	if p.Title == "" {
		return ErrTitleRequired
	}
	if p.Content == "" {
		return ErrContentRequired
	}
	return nil
}
