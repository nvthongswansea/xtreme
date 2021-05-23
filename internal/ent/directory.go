// Code generated by entc, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/nvthongswansea/xtreme/internal/ent/directory"
	"github.com/nvthongswansea/xtreme/internal/ent/user"
)

// Directory is the model entity for the Directory schema.
type Directory struct {
	config `json:"-"`
	// ID of the ent.
	ID string `json:"id,omitempty"`
	// Name holds the value of the "name" field.
	Name string `json:"name,omitempty"`
	// Path holds the value of the "path" field.
	Path string `json:"path,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the DirectoryQuery when eager-loading is set.
	Edges                DirectoryEdges `json:"edges"`
	directory_child_dirs *string
	user_directories     *string
}

// DirectoryEdges holds the relations/edges for other nodes in the graph.
type DirectoryEdges struct {
	// Owner holds the value of the owner edge.
	Owner *User `json:"owner,omitempty"`
	// ChildFiles holds the value of the childFiles edge.
	ChildFiles []*File `json:"childFiles,omitempty"`
	// Parent holds the value of the parent edge.
	Parent *Directory `json:"parent,omitempty"`
	// ChildDirs holds the value of the childDirs edge.
	ChildDirs []*Directory `json:"childDirs,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [4]bool
}

// OwnerOrErr returns the Owner value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e DirectoryEdges) OwnerOrErr() (*User, error) {
	if e.loadedTypes[0] {
		if e.Owner == nil {
			// The edge owner was loaded in eager-loading,
			// but was not found.
			return nil, &NotFoundError{label: user.Label}
		}
		return e.Owner, nil
	}
	return nil, &NotLoadedError{edge: "owner"}
}

// ChildFilesOrErr returns the ChildFiles value or an error if the edge
// was not loaded in eager-loading.
func (e DirectoryEdges) ChildFilesOrErr() ([]*File, error) {
	if e.loadedTypes[1] {
		return e.ChildFiles, nil
	}
	return nil, &NotLoadedError{edge: "childFiles"}
}

// ParentOrErr returns the Parent value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e DirectoryEdges) ParentOrErr() (*Directory, error) {
	if e.loadedTypes[2] {
		if e.Parent == nil {
			// The edge parent was loaded in eager-loading,
			// but was not found.
			return nil, &NotFoundError{label: directory.Label}
		}
		return e.Parent, nil
	}
	return nil, &NotLoadedError{edge: "parent"}
}

// ChildDirsOrErr returns the ChildDirs value or an error if the edge
// was not loaded in eager-loading.
func (e DirectoryEdges) ChildDirsOrErr() ([]*Directory, error) {
	if e.loadedTypes[3] {
		return e.ChildDirs, nil
	}
	return nil, &NotLoadedError{edge: "childDirs"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Directory) scanValues(columns []string) ([]interface{}, error) {
	values := make([]interface{}, len(columns))
	for i := range columns {
		switch columns[i] {
		case directory.FieldID, directory.FieldName, directory.FieldPath:
			values[i] = new(sql.NullString)
		case directory.FieldCreatedAt, directory.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		case directory.ForeignKeys[0]: // directory_child_dirs
			values[i] = new(sql.NullString)
		case directory.ForeignKeys[1]: // user_directories
			values[i] = new(sql.NullString)
		default:
			return nil, fmt.Errorf("unexpected column %q for type Directory", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Directory fields.
func (d *Directory) assignValues(columns []string, values []interface{}) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case directory.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				d.ID = value.String
			}
		case directory.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				d.Name = value.String
			}
		case directory.FieldPath:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field path", values[i])
			} else if value.Valid {
				d.Path = value.String
			}
		case directory.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				d.CreatedAt = value.Time
			}
		case directory.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				d.UpdatedAt = value.Time
			}
		case directory.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field directory_child_dirs", values[i])
			} else if value.Valid {
				d.directory_child_dirs = new(string)
				*d.directory_child_dirs = value.String
			}
		case directory.ForeignKeys[1]:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field user_directories", values[i])
			} else if value.Valid {
				d.user_directories = new(string)
				*d.user_directories = value.String
			}
		}
	}
	return nil
}

// QueryOwner queries the "owner" edge of the Directory entity.
func (d *Directory) QueryOwner() *UserQuery {
	return (&DirectoryClient{config: d.config}).QueryOwner(d)
}

// QueryChildFiles queries the "childFiles" edge of the Directory entity.
func (d *Directory) QueryChildFiles() *FileQuery {
	return (&DirectoryClient{config: d.config}).QueryChildFiles(d)
}

// QueryParent queries the "parent" edge of the Directory entity.
func (d *Directory) QueryParent() *DirectoryQuery {
	return (&DirectoryClient{config: d.config}).QueryParent(d)
}

// QueryChildDirs queries the "childDirs" edge of the Directory entity.
func (d *Directory) QueryChildDirs() *DirectoryQuery {
	return (&DirectoryClient{config: d.config}).QueryChildDirs(d)
}

// Update returns a builder for updating this Directory.
// Note that you need to call Directory.Unwrap() before calling this method if this Directory
// was returned from a transaction, and the transaction was committed or rolled back.
func (d *Directory) Update() *DirectoryUpdateOne {
	return (&DirectoryClient{config: d.config}).UpdateOne(d)
}

// Unwrap unwraps the Directory entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (d *Directory) Unwrap() *Directory {
	tx, ok := d.config.driver.(*txDriver)
	if !ok {
		panic("ent: Directory is not a transactional entity")
	}
	d.config.driver = tx.drv
	return d
}

// String implements the fmt.Stringer.
func (d *Directory) String() string {
	var builder strings.Builder
	builder.WriteString("Directory(")
	builder.WriteString(fmt.Sprintf("id=%v", d.ID))
	builder.WriteString(", name=")
	builder.WriteString(d.Name)
	builder.WriteString(", path=")
	builder.WriteString(d.Path)
	builder.WriteString(", created_at=")
	builder.WriteString(d.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", updated_at=")
	builder.WriteString(d.UpdatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// Directories is a parsable slice of Directory.
type Directories []*Directory

func (d Directories) config(cfg config) {
	for _i := range d {
		d[_i].config = cfg
	}
}