package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// PleasePay holds the schema definition for the PayRequest entity.
type PleasePay struct {
	ent.Schema
}

// Fields of the PleasePay.
func (PleasePay) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Float("amount"),
		field.String("state").Default("PENDING"),
	}
}

// Edges of the PleasePay.
func (PleasePay) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("transaction", Transaction.Type).Unique(),
		edge.To("receiver", User.Type).Unique(),
	}
}
