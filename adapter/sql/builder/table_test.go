package builder

import (
	"testing"
	"time"

	"github.com/jiyeuran/dbm"
	"github.com/jiyeuran/dbm/adapter/sql"
	"github.com/stretchr/testify/assert"
)

func TestTable_Build(t *testing.T) {
	var (
		tableBuilder = Table{
			BufferFactory: BufferFactory{InlineValues: true, BoolTrueValue: "true", BoolFalseValue: "false", Quoter: Quote{IDPrefix: "`", IDSuffix: "`", IDSuffixEscapeChar: "`", ValueQuote: "'", ValueQuoteEscapeChar: "'"}},
			ColumnMapper:  sql.ColumnMapper,
			DropKeyMapper: sql.DropKeyMapper,
		}
	)

	tests := []struct {
		result string
		table  dbm.Table
	}{
		{
			result: "CREATE TABLE `products` (`id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY, `name` VARCHAR(255), `description` TEXT);",
			table: dbm.Table{
				Op:   dbm.SchemaCreate,
				Name: "products",
				Definitions: []dbm.TableDefinition{
					dbm.Column{Name: "id", Type: dbm.ID, Primary: true},
					dbm.Column{Name: "name", Type: dbm.String},
					dbm.Column{Name: "description", Type: dbm.Text},
				},
			},
		},
		{
			result: "CREATE TABLE `products_2021` PARTITION OF `products` FOR VALUES FROM ('2021-01-01') TO ('2021-01-31');",
			table: dbm.Table{
				Op:          dbm.SchemaCreate,
				Name:        "products_2021",
				Definitions: []dbm.TableDefinition{},
				Options:     "PARTITION OF `products` FOR VALUES FROM ('2021-01-01') TO ('2021-01-31')",
			},
		},
		{
			result: "CREATE TABLE `columns` (`bool` BOOL NOT NULL DEFAULT false, `int` INT(11) UNSIGNED, `bigint` BIGINT(20) UNSIGNED, `float` FLOAT(24) UNSIGNED, `decimal` DECIMAL(6,2) UNSIGNED, `string` VARCHAR(144) UNIQUE, `text` TEXT(1000), `date` DATE, `datetime` DATETIME DEFAULT '2020-01-01 01:00:00', `time` TIME, `blob` blob, PRIMARY KEY (`int`), FOREIGN KEY (`int`, `string`) REFERENCES `products` (`id`, `name`) ON DELETE CASCADE ON UPDATE CASCADE, UNIQUE `date_unique` (`date`)) Engine=InnoDB;",
			table: dbm.Table{
				Op:   dbm.SchemaCreate,
				Name: "columns",
				Definitions: []dbm.TableDefinition{
					dbm.Column{Name: "bool", Type: dbm.Bool, Required: true, Default: false},
					dbm.Column{Name: "int", Type: dbm.Int, Limit: 11, Unsigned: true},
					dbm.Column{Name: "bigint", Type: dbm.BigInt, Limit: 20, Unsigned: true},
					dbm.Column{Name: "float", Type: dbm.Float, Precision: 24, Unsigned: true},
					dbm.Column{Name: "decimal", Type: dbm.Decimal, Precision: 6, Scale: 2, Unsigned: true},
					dbm.Column{Name: "string", Type: dbm.String, Limit: 144, Unique: true},
					dbm.Column{Name: "text", Type: dbm.Text, Limit: 1000},
					dbm.Column{Name: "date", Type: dbm.Date},
					dbm.Column{Name: "datetime", Type: dbm.DateTime, Default: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)},
					dbm.Column{Name: "time", Type: dbm.Time},
					dbm.Column{Name: "blob", Type: "blob"},
					dbm.Key{Columns: []string{"int"}, Type: dbm.PrimaryKey},
					dbm.Key{Columns: []string{"int", "string"}, Type: dbm.ForeignKey, Reference: dbm.ForeignKeyReference{Table: "products", Columns: []string{"id", "name"}, OnDelete: "CASCADE", OnUpdate: "CASCADE"}},
					dbm.Key{Columns: []string{"date"}, Name: "date_unique", Type: dbm.UniqueKey},
				},
				Options: "Engine=InnoDB",
			},
		},
		{
			result: "CREATE TABLE IF NOT EXISTS `products` (`id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, `data` TEXT, `raw` BOOL);",
			table: dbm.Table{
				Op:       dbm.SchemaCreate,
				Name:     "products",
				Optional: true,
				Definitions: []dbm.TableDefinition{
					dbm.Column{Name: "id", Type: dbm.BigID, Primary: true},
					dbm.Column{Name: "data", Type: dbm.JSON},
					dbm.Raw("`raw` BOOL"),
				},
			},
		},
		{
			result: "ALTER TABLE `columns` ADD COLUMN `verified` BOOL;ALTER TABLE `columns` RENAME COLUMN `string` TO `name`;ALTER TABLE `columns` ;ALTER TABLE `columns` DROP COLUMN `blob`;",
			table: dbm.Table{
				Op:   dbm.SchemaAlter,
				Name: "columns",
				Definitions: []dbm.TableDefinition{
					dbm.Column{Name: "verified", Type: dbm.Bool, Op: dbm.SchemaCreate},
					dbm.Column{Name: "string", Rename: "name", Op: dbm.SchemaRename},
					dbm.Column{Name: "bool", Type: dbm.Int, Op: dbm.SchemaAlter},
					dbm.Column{Name: "blob", Op: dbm.SchemaDrop},
				},
			},
		},
		{
			result: "ALTER TABLE `transactions` ADD FOREIGN KEY (`user_id`) REFERENCES `products` (`id`, `name`) ON DELETE CASCADE ON UPDATE CASCADE;",
			table: dbm.Table{
				Op:   dbm.SchemaAlter,
				Name: "transactions",
				Definitions: []dbm.TableDefinition{
					dbm.Key{Columns: []string{"user_id"}, Type: dbm.ForeignKey, Reference: dbm.ForeignKeyReference{Table: "products", Columns: []string{"id", "name"}, OnDelete: "CASCADE", OnUpdate: "CASCADE"}},
				},
			},
		},
		{
			result: "ALTER TABLE `transactions` DROP CONSTRAINT `fk`;",
			table: dbm.Table{
				Op:   dbm.SchemaAlter,
				Name: "transactions",
				Definitions: []dbm.TableDefinition{
					dbm.Key{Op: dbm.SchemaDrop, Name: "fk", Type: dbm.ForeignKey},
				},
			},
		},
		{
			result: "ALTER TABLE `table` RENAME TO `table1`;",
			table: dbm.Table{
				Op:     dbm.SchemaRename,
				Name:   "table",
				Rename: "table1",
			},
		},
		{
			result: "DROP TABLE `table`;",
			table: dbm.Table{
				Op:   dbm.SchemaDrop,
				Name: "table",
			},
		},
		{
			result: "DROP TABLE IF EXISTS `table`;",
			table: dbm.Table{
				Op:       dbm.SchemaDrop,
				Name:     "table",
				Optional: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.result, func(t *testing.T) {
			assert.Equal(t, test.result, tableBuilder.Build(test.table))
		})
	}
}

func TestTable_BuildWithDefinitionFilter(t *testing.T) {
	var (
		definitionFilter = func(table dbm.Table, def dbm.TableDefinition) bool {
			_, ok := def.(dbm.Key)
			// https://www.sqlite.org/omitted.html
			// > Only the RENAME TABLE, ADD COLUMN, RENAME COLUMN, and DROP COLUMN variants of the ALTER TABLE command are supported.
			if ok && table.Op == dbm.SchemaAlter {
				return false
			}

			return true
		}
		tableBuilder = Table{
			BufferFactory:    BufferFactory{InlineValues: true, BoolTrueValue: "true", BoolFalseValue: "false", Quoter: Quote{IDPrefix: "`", IDSuffix: "`", IDSuffixEscapeChar: "`", ValueQuote: "'", ValueQuoteEscapeChar: "'"}},
			ColumnMapper:     sql.ColumnMapper,
			DefinitionFilter: definitionFilter,
		}
	)

	tests := []struct {
		result string
		table  dbm.Table
	}{
		{
			result: "CREATE TABLE `columns` (`bool` BOOL NOT NULL DEFAULT false, `int` INT(11) UNSIGNED, `bigint` BIGINT(20) UNSIGNED, `float` FLOAT(24) UNSIGNED, `decimal` DECIMAL(6,2) UNSIGNED, `string` VARCHAR(144) UNIQUE, `text` TEXT(1000), `date` DATE, `datetime` DATETIME DEFAULT '2020-01-01 01:00:00', `time` TIME, `blob` blob, PRIMARY KEY (`int`), FOREIGN KEY (`int`, `string`) REFERENCES `products` (`id`, `name`) ON DELETE CASCADE ON UPDATE CASCADE, UNIQUE `date_unique` (`date`)) Engine=InnoDB;",
			table: dbm.Table{
				Op:   dbm.SchemaCreate,
				Name: "columns",
				Definitions: []dbm.TableDefinition{
					dbm.Column{Name: "bool", Type: dbm.Bool, Required: true, Default: false},
					dbm.Column{Name: "int", Type: dbm.Int, Limit: 11, Unsigned: true},
					dbm.Column{Name: "bigint", Type: dbm.BigInt, Limit: 20, Unsigned: true},
					dbm.Column{Name: "float", Type: dbm.Float, Precision: 24, Unsigned: true},
					dbm.Column{Name: "decimal", Type: dbm.Decimal, Precision: 6, Scale: 2, Unsigned: true},
					dbm.Column{Name: "string", Type: dbm.String, Limit: 144, Unique: true},
					dbm.Column{Name: "text", Type: dbm.Text, Limit: 1000},
					dbm.Column{Name: "date", Type: dbm.Date},
					dbm.Column{Name: "datetime", Type: dbm.DateTime, Default: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)},
					dbm.Column{Name: "time", Type: dbm.Time},
					dbm.Column{Name: "blob", Type: "blob"},
					dbm.Key{Columns: []string{"int"}, Type: dbm.PrimaryKey},
					dbm.Key{Columns: []string{"int", "string"}, Type: dbm.ForeignKey, Reference: dbm.ForeignKeyReference{Table: "products", Columns: []string{"id", "name"}, OnDelete: "CASCADE", OnUpdate: "CASCADE"}},
					dbm.Key{Columns: []string{"date"}, Name: "date_unique", Type: dbm.UniqueKey},
				},
				Options: "Engine=InnoDB",
			},
		},
		{
			result: "ALTER TABLE `columns` ADD COLUMN `verified` BOOL;ALTER TABLE `columns` RENAME COLUMN `string` TO `name`;ALTER TABLE `columns` ;ALTER TABLE `columns` DROP COLUMN `blob`;",
			table: dbm.Table{
				Op:   dbm.SchemaAlter,
				Name: "columns",
				Definitions: []dbm.TableDefinition{
					dbm.Column{Name: "verified", Type: dbm.Bool, Op: dbm.SchemaCreate},
					dbm.Column{Name: "string", Rename: "name", Op: dbm.SchemaRename},
					dbm.Column{Name: "bool", Type: dbm.Int, Op: dbm.SchemaAlter},
					dbm.Column{Name: "blob", Op: dbm.SchemaDrop},
				},
			},
		},
		{
			result: "",
			table: dbm.Table{
				Op:   dbm.SchemaAlter,
				Name: "transactions",
				Definitions: []dbm.TableDefinition{
					dbm.Key{Op: dbm.SchemaCreate, Columns: []string{"user_id"}, Type: dbm.ForeignKey, Reference: dbm.ForeignKeyReference{Table: "products", Columns: []string{"id", "name"}, OnDelete: "CASCADE", OnUpdate: "CASCADE"}},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.result, func(t *testing.T) {
			assert.Equal(t, test.result, tableBuilder.Build(test.table))
		})
	}
}
