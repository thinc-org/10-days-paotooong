package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("email"),
		field.String("hash"),
		field.String("first_name"),
		field.String("family_name"),
		field.Float("money"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
