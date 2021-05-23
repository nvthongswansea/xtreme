package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").NotEmpty().Unique().Immutable(),
		field.String("username").NotEmpty().Unique().Immutable(),
		field.String("password_hash").NotEmpty(),
		field.Time("created_at").Immutable().Default(func() time.Time {
			return time.Now()
		}),
		field.Time("updated_at").
			Default(func() time.Time {
				return time.Now()
			}).
			UpdateDefault(func() time.Time {
				return time.Now()
			}),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("files", File.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		edge.To("directories", Directory.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}
