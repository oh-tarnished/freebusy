// Package runtime assembles the freebusy gRPC services from configuration: it
// opens the configured database backend, builds the provider-agnostic
// repository, and constructs the protobuf service implementations in the sibling
// packages (runtime/promocode). The transport layer (package internal) registers
// what this package builds; the database layer stays agnostic to protobuf.
package runtime

import (
	"github.com/oh-tarnished/freebusy/internal/database"
	"github.com/oh-tarnished/freebusy/internal/runtime/promocode"
	promocodedb "github.com/oh-tarnished/freebusy/internal/service/promocode/db"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
)

// NewPromoCodeServer opens the configured backend, builds the repository, and
// returns the promocode gRPC service implementation ready to register.
func NewPromoCodeServer() (promocodepbv1.PromoCodeServiceServer, error) {
	conn, err := database.Open()
	if err != nil {
		return nil, err
	}
	return promocode.NewServer(promocodedb.New(conn)), nil
}

// Other Services can be added here in the future, following the same pattern: open the database connection, build the repository, and return the service implementation.
