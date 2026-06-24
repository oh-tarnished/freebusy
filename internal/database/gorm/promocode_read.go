package gorm

import (
	"errors"

	"github.com/oh-tarnished/freebusy/internal/types"
	"gorm.io/gorm"
)

// mapGormErr translates GORM sentinels into repository sentinels so the service
// layer stays free of storage-specific error types.
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
