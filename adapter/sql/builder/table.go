package builder

import (
	"log"
	"strconv"

	"github.com/jiyeuran/dbm"
)

type ColumnMapper func(*dbm.Column) (string, int, int)
type DropKeyMapper func(dbm.KeyType) string
type DefinitionFilter func(table dbm.Table, def dbm.TableDefinition) bool

// Table builder.
type Table struct {
	BufferFactory    BufferFactory
	ColumnMapper     ColumnMapper
	DropKeyMapper    DropKeyMapper
	DefinitionFilter DefinitionFilter
}

// Build SQL query for table creation and modification.
func (t Table) Build(table dbm.Table) string {
	var (
		buffer = t.BufferFactory.Create()
	)

	switch table.Op {
	case dbm.SchemaCreate:
		t.WriteCreateTable(&buffer, table)
	case dbm.SchemaAlter:
		t.WriteAlterTable(&buffer, table)
	case dbm.SchemaRename:
		t.WriteRenameTable(&buffer, table)
	case dbm.SchemaDrop:
		t.WriteDropTable(&buffer, table)
	}

	return buffer.String()
}

// WriteCreateTable query to buffer.
func (t Table) WriteCreateTable(buffer *Buffer, table dbm.Table) {
	defs := t.definitions(table)

	buffer.WriteString("CREATE TABLE ")

	if table.Optional {
		buffer.WriteString("IF NOT EXISTS ")
	}

	buffer.WriteEscape(table.Name)
	if len(defs) > 0 {
		buffer.WriteString(" (")

		for i, def := range defs {
			if i > 0 {
				buffer.WriteString(", ")
			}
			switch v := def.(type) {
			case dbm.Column:
				t.WriteColumn(buffer, v)
			case dbm.Key:
				t.WriteKey(buffer, v)
			case dbm.Raw:
				buffer.WriteString(string(v))
			}
		}

		buffer.WriteByte(')')
	}
	t.WriteOptions(buffer, table.Options)
	buffer.WriteByte(';')
}

// WriteAlterTable query to buffer.
func (t Table) WriteAlterTable(buffer *Buffer, table dbm.Table) {
	defs := t.definitions(table)

	for _, def := range defs {
		buffer.WriteString("ALTER TABLE ")
		buffer.WriteEscape(table.Name)
		buffer.WriteByte(' ')

		switch v := def.(type) {
		case dbm.Column:
			switch v.Op {
			case dbm.SchemaCreate:
				buffer.WriteString("ADD COLUMN ")
				t.WriteColumn(buffer, v)
			case dbm.SchemaRename:
				// Add Change
				buffer.WriteString("RENAME COLUMN ")
				buffer.WriteEscape(v.Name)
				buffer.WriteString(" TO ")
				buffer.WriteEscape(v.Rename)
			case dbm.SchemaDrop:
				buffer.WriteString("DROP COLUMN ")
				buffer.WriteEscape(v.Name)
			}
		case dbm.Key:
			// TODO: Rename and Drop, PR welcomed.
			switch v.Op {
			case dbm.SchemaCreate:
				buffer.WriteString("ADD ")
				t.WriteKey(buffer, v)
			case dbm.SchemaDrop:
				buffer.WriteString("DROP ")
				buffer.WriteString(t.DropKeyMapper(v.Type))
				buffer.WriteString(" ")
				buffer.WriteEscape(v.Name)
			}
		}

		t.WriteOptions(buffer, table.Options)
		buffer.WriteByte(';')
	}
}

// WriteRenameTable query to buffer.
func (t Table) WriteRenameTable(buffer *Buffer, table dbm.Table) {
	buffer.WriteString("ALTER TABLE ")
	buffer.WriteEscape(table.Name)
	buffer.WriteString(" RENAME TO ")
	buffer.WriteEscape(table.Rename)
	buffer.WriteByte(';')
}

// WriteDropTable query to buffer.
func (t Table) WriteDropTable(buffer *Buffer, table dbm.Table) {
	buffer.WriteString("DROP TABLE ")

	if table.Optional {
		buffer.WriteString("IF EXISTS ")
	}

	buffer.WriteEscape(table.Name)
	buffer.WriteByte(';')
}

// WriteColumn definition to buffer.
func (t Table) WriteColumn(buffer *Buffer, column dbm.Column) {
	var (
		typ, m, n = t.ColumnMapper(&column)
	)

	buffer.WriteEscape(column.Name)
	buffer.WriteByte(' ')
	buffer.WriteString(typ)

	if m != 0 {
		buffer.WriteByte('(')
		buffer.WriteString(strconv.Itoa(m))

		if n != 0 {
			buffer.WriteByte(',')
			buffer.WriteString(strconv.Itoa(n))
		}

		buffer.WriteByte(')')
	}

	if column.Unsigned {
		buffer.WriteString(" UNSIGNED")
	}

	if column.Unique {
		buffer.WriteString(" UNIQUE")
	}

	if column.Required {
		buffer.WriteString(" NOT NULL")
	}

	if column.Primary {
		buffer.WriteString(" PRIMARY KEY")
	}

	if column.Default != nil {
		buffer.WriteString(" DEFAULT ")
		buffer.WriteValue(column.Default)
	}

	t.WriteOptions(buffer, column.Options)
}

// WriteKey definition to buffer.
func (t Table) WriteKey(buffer *Buffer, key dbm.Key) {
	var (
		typ = string(key.Type)
	)

	buffer.WriteString(typ)

	if key.Name != "" {
		buffer.WriteByte(' ')
		buffer.WriteEscape(key.Name)
	}

	buffer.WriteString(" (")
	for i, col := range key.Columns {
		if i > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteEscape(col)
	}
	buffer.WriteString(")")

	if key.Type == dbm.ForeignKey {
		buffer.WriteString(" REFERENCES ")
		buffer.WriteEscape(key.Reference.Table)

		buffer.WriteString(" (")
		for i, col := range key.Reference.Columns {
			if i > 0 {
				buffer.WriteString(", ")
			}
			buffer.WriteEscape(col)
		}
		buffer.WriteString(")")

		if onDelete := key.Reference.OnDelete; onDelete != "" {
			buffer.WriteString(" ON DELETE ")
			buffer.WriteString(onDelete)
		}

		if onUpdate := key.Reference.OnUpdate; onUpdate != "" {
			buffer.WriteString(" ON UPDATE ")
			buffer.WriteString(onUpdate)
		}
	}

	t.WriteOptions(buffer, key.Options)
}

// WriteOptions sql to buffer.
func (t Table) WriteOptions(buffer *Buffer, options string) {
	if options == "" {
		return
	}

	buffer.WriteByte(' ')
	buffer.WriteString(options)
}

func (t Table) definitions(table dbm.Table) []dbm.TableDefinition {
	if t.DefinitionFilter == nil {
		return table.Definitions
	}

	result := []dbm.TableDefinition{}

	for _, def := range table.Definitions {
		if t.DefinitionFilter(table, def) {
			result = append(result, def)
		} else {
			log.Printf("[DBM] An unsupported table definition has been excluded: %T", def)
		}
	}

	return result
}
