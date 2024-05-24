package rule

import (
	"context"
	"entgo.io/ent/entql"
	"github.com/suyuan32/simple-admin-common/orm/ent/tenantctx"
	"simpleTenant/ent/privacy"
)

// FilterTenantRule is a query/mutation rule that filters out entities that are not in the tenant.
func FilterTenantRule() privacy.QueryMutationRule {
	// TenantsFilter is an interface to wrap WhereTenantID()
	// predicate that is used by both `Group` and `User` schemas.
	type TenantsFilter interface {
		WhereTenantID(entql.IntP)
	}
	return privacy.FilterFunc(func(ctx context.Context, f privacy.Filter) error {
		tenantId, err := tenantctx.GetTenantIDFromCtx(ctx)
		if err != nil {
			return privacy.Denyf("%v", err)
		}

		// use 0 as admin
		if tenantId == 0 {
			return privacy.Allow
		}

		tf, ok := f.(TenantsFilter)
		if !ok {
			return privacy.Denyf("unexpected filter type %T", f)
		}
		// Make sure that a tenant reads only entities that have an edge to it.
		tf.WhereTenantID(entql.IntEQ(tenantId))
		// Skip to the next privacy rule (equivalent to return nil).
		return privacy.Skip
	})
}
