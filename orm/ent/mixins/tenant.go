package mixins

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/suyuan32/simple-admin-common/orm/ent/rule"
)

// TenantMixin for embedding the tenant info in different schemas.
type TenantMixin struct {
	mixin.Schema
}

// Fields for all schemas that embed TenantMixin.
func (TenantMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("tenant_id").
			Immutable(),
	}
}

// Policy for all schemas that embed TenantMixin.
func (TenantMixin) Policy() ent.Policy {
	return rule.FilterTenantRule()
}
