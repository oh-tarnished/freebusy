package gorm

import (
	"errors"

	"github.com/oh-tarnished/freebusy/internal/types"
	"gorm.io/gorm"
)

// mapGormErr translates GORM sentinel errors into the provider-neutral errors in
// internal/types, leaving anything else untouched for the caller to wrap.
func mapGormErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return types.ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return types.ErrAlreadyExists
	default:
		return err
	}
}
