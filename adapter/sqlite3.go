package adapter

import (
	"log"
	"strings"

	"github.com/jiyeuran/dbm"
	"github.com/jiyeuran/dbm/adapter/sql"
)

type sqlite3 struct{}

func (sqlite3) errorMapper(err error) error {
	if err == nil {
		return nil
	}

	var (
		msg         = err.Error()
		failedSep   = " failed: "
		failedIndex = strings.Index(msg, failedSep)
		failedLen   = 9 // len(failedSep)
	)

	if failedIndex < 0 {
		failedIndex = 0
	}

	switch msg[:failedIndex] {
	case "UNIQUE constraint":
		return dbm.ConstraintError{
			Key:  msg[failedIndex+failedLen:],
			Type: dbm.UniqueConstraint,
			Err:  err,
		}
	case "CHECK constraint":
		return dbm.ConstraintError{
			Key:  msg[failedIndex+failedLen:],
			Type: dbm.CheckConstraint,
			Err:  err,
		}
	default:
		return err
	}
}

func (sqlite3) columnMapper(column *dbm.Column) (string, int, int) {
	var (
		typ      string
		m, n     int
		unsigned = column.Unsigned
	)

	column.Unsigned = false

	switch column.Type {
	case dbm.ID:
		typ = "INTEGER"
	case dbm.BigID:
		typ = "BIGINT"
	case dbm.Int:
		typ = "INTEGER"
		m = column.Limit
	default:
		typ, m, n = sql.ColumnMapper(column)
	}

	if unsigned {
		typ = "UNSIGNED " + typ
	}

	return typ, m, n
}

func (sqlite3) definitionFilter(table dbm.Table, def dbm.TableDefinition) bool {
	if table.Op == dbm.SchemaAlter {
		// https://www.sqlite.org/omitted.html
		// > Only the RENAME TABLE, ADD COLUMN, RENAME COLUMN, and DROP COLUMN variants of the ALTER TABLE command are supported.
		_, ok := def.(dbm.Key)
		if ok {
			log.Print("[DBM] SQLite3 adapter does not support adding keys when modifying tables")

			return false
		}
	}

	return true
}
