package mixins

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/suyuan32/simple-admin-common/orm/ent/entenum"
)

// TenantMixin for embedding the tenant info in different schemas.
type TenantMixin struct {
	mixin.Schema
}

// Fields for all schemas that embed TenantMixin.
func (TenantMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("tenant_id").
			Default(entenum.TenantDefaultId).
			Immutable().Comment("Tenant ID | 租户 ID"),
	}
}
