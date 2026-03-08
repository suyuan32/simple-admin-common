package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CasbinRule holds the schema definition for the CasbinRule entity.
type CasbinRule struct {
	ent.Schema
}

// Fields of the CasbinRule.
func (CasbinRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("Ptype").MaxLen(24).Default(""),
		field.String("V0").MaxLen(24).Default(""),
		field.String("V1").Default(""),
		field.String("V2").MaxLen(24).Default(""),
		field.String("V3").MaxLen(32).Default(""),
		field.String("V4").MaxLen(12).Default(""),
		field.String("V5").MaxLen(12).Default(""),
	}
}

// Edges of the CasbinRule.
func (CasbinRule) Edges() []ent.Edge {
	return nil
}

func (CasbinRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("Ptype", "V0", "V1", "V2", "V3", "V4", "V5").Unique(),
	}
}
