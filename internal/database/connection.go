package database

import (
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	"gorm.io/gorm"
)

type Connection struct {
	PgSQLConn *gorm.DB
	Hasura    *freebusyql.Service
}
