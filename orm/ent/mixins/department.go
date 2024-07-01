package mixins

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// DepartmentMixin for embedding the department info in different schemas.
type DepartmentMixin struct {
	mixin.Schema
}

// Fields for all schemas that embed DepartmentMixin.
func (DepartmentMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("department_id").
			Optional().
			Comment("Department ID | 部门 ID"),
	}
}
