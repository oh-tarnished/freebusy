// Money row lookup shared by unit hydration.
package hasura

import (
	"context"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/moneysql"
	"github.com/oh-tarnished/freebusy/internal/service/dbutil"
)

func (r *PropertyRepository) money(ctx context.Context, id string) (*moneysql.CommonMoneys, error) {
	m, err := r.svc.Query.Common.Moneys.Get(ctx, id)
	if err != nil {
		return nil, dbutil.MapHasuraErr(err)
	}
	return m, nil
}
