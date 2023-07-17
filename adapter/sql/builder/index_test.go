package builder

import (
	"testing"

	"github.com/jiyeuran/dbm"
	"github.com/stretchr/testify/assert"
)

func TestIndex_Build(t *testing.T) {
	var (
		bufferFactory = BufferFactory{ArgumentPlaceholder: "?", InlineValues: true, BoolTrueValue: "true", BoolFalseValue: "false", Quoter: Quote{IDPrefix: "`", IDSuffix: "`", IDSuffixEscapeChar: "`", ValueQuote: "'", ValueQuoteEscapeChar: "'"}}
		indexBuilder  = Index{
			BufferFactory:    bufferFactory,
			DropIndexOnTable: true,
		}
	)

	tests := []struct {
		result string
		index  dbm.Index
	}{
		{
			result: "CREATE INDEX `index` ON `table` (`column1`);",
			index: dbm.Index{
				Op:      dbm.SchemaCreate,
				Table:   "table",
				Name:    "index",
				Columns: []string{"column1"},
			},
		},
		{
			result: "CREATE UNIQUE INDEX `index` ON `table` (`column1`);",
			index: dbm.Index{
				Op:      dbm.SchemaCreate,
				Table:   "table",
				Name:    "index",
				Unique:  true,
				Columns: []string{"column1"},
			},
		},
		{
			result: "CREATE INDEX `index` ON `table` (`column1`, `column2`);",
			index: dbm.Index{
				Op:      dbm.SchemaCreate,
				Table:   "table",
				Name:    "index",
				Columns: []string{"column1", "column2"},
			},
		},
		{
			result: "CREATE INDEX IF NOT EXISTS `index` ON `table` (`column1`);",
			index: dbm.Index{
				Op:       dbm.SchemaCreate,
				Table:    "table",
				Name:     "index",
				Optional: true,
				Columns:  []string{"column1"},
			},
		},
		{
			result: "CREATE INDEX IF NOT EXISTS `index` ON `table` (`column1`) COMMENT 'comment';",
			index: dbm.Index{
				Op:       dbm.SchemaCreate,
				Table:    "table",
				Name:     "index",
				Optional: true,
				Columns:  []string{"column1"},
				Options:  "COMMENT 'comment'",
			},
		},
		{
			result: "DROP INDEX `index` ON `table`;",
			index: dbm.Index{
				Op:    dbm.SchemaDrop,
				Name:  "index",
				Table: "table",
			},
		},
		{
			result: "DROP INDEX IF EXISTS `index` ON `table`;",
			index: dbm.Index{
				Op:       dbm.SchemaDrop,
				Name:     "index",
				Table:    "table",
				Optional: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.result, func(t *testing.T) {
			assert.Equal(t, test.result, indexBuilder.Build(test.index))
		})
	}
}
