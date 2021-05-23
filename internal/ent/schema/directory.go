package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"time"
)

// Directory holds the schema definition for the Directory entity.
type Directory struct {
	ent.Schema
}

// Fields of the Directory.
func (Directory) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").NotEmpty().Unique().Immutable(),
		field.String("name").NotEmpty(),
		field.String("path"),
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

// Edges of the Directory.
func (Directory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).Ref("directories").Unique().Required(),
		edge.To("childFiles", File.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		edge.To("childDirs", Directory.Type).From("parent").Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (Directory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").
			Edges("parent").
			Unique(),
	}
}
