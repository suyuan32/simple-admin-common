package mixins

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/gofrs/uuid/v5"
)

// CreatedByMixin for embedding the created user's uuid info in different schemas.
type CreatedByMixin struct {
	mixin.Schema
}

// Fields for all schemas that embed CreatedByMixin.
func (CreatedByMixin) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("created_by", uuid.UUID{}).
			Optional().
			Comment("Created user's UUID | 创建者 UUID"),
	}
}
