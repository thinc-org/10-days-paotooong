package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Transaction holds the schema definition for the Transaction entity.
type Transaction struct {
	ent.Schema
}

// Fields of the Transaction.
func (Transaction) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("payer_id", uuid.UUID{}),
		field.UUID("receiver_id", uuid.UUID{}),
		field.Float("amount"),
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the Transaction.
func (Transaction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("payer", User.Type).
			Ref("out_transactions").
			Unique().
			Field("payer_id").
			Required(),
		edge.From("receiver", User.Type).
			Ref("in_transactions").
			Unique().
			Field("receiver_id").
			Required(),
	}
}
