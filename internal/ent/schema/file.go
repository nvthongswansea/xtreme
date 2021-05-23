package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"time"
)

// File holds the schema definition for the File entity.
type File struct {
	ent.Schema
}

// Fields of the File.
func (File) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").NotEmpty().Unique().Immutable(),
		field.String("name").NotEmpty(),
		field.String("mime_type").Default("unknown"),
		field.String("path").Default(""),
		field.String("rel_path_on_disk").NotEmpty().Unique(),
		field.Int64("size").NonNegative(),
		field.Bool("is_deleted").Default(false),
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

// Edges of the File.
func (File) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).Ref("files").Unique().Required(),
		edge.From("parent", Directory.Type).Ref("childFiles").Unique().Required(),
	}
}
func (File) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").
			Edges("parent").
			Unique(),
	}
}
