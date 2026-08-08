package service

import (
	"errors"

	"github.com/mohdrashid9678/xlink/internal/repository"
)

var (
	ErrNotFound   = errors.New("URL not found")
	ErrConflict   = errors.New("short code is already in use")
	ErrValidation = errors.New("invalid URL input")
)

func mapRepositoryError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return ErrNotFound
	case errors.Is(err, repository.ErrConflict):
		return ErrConflict
	default:
		return err
	}
}
