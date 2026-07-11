// Money row lookup shared by unit hydration.
package hasura

import (
	"context"
	"github.com/oh-tarnished/freebusy/internal/service/dbutil"

	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
)

func (r *PropertyRepository) money(ctx context.Context, id string) (*commonschema.CommonMoneys, error) {
	m, err := r.svc.Query.Common.Moneys.Get(ctx, id)
	if err != nil {
		return nil, dbutil.MapHasuraErr(err)
	}
	return m, nil
}
