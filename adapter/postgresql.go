package adapter

import (
	"database/sql/driver"
	"strings"
	"time"

	"github.com/jiyeyuran/dbm"
	"github.com/jiyeyuran/dbm/adapter/sql"
)

type postgres struct{}

func (q postgres) ID(name string) string {
	end := strings.IndexRune(name, 0)
	if end > -1 {
		name = name[:end]
	}
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func (q postgres) Value(v interface{}) string {
	switch v := v.(type) {
	default:
		panic("unsupported value")
	case string:
		v = strings.ReplaceAll(v, `'`, `''`)
		if strings.Contains(v, `\`) {
			v = strings.ReplaceAll(v, `\`, `\\`)
			v = ` E'` + v + `'`
		} else {
			v = `'` + v + `'`
		}
		return v
	}
}

func (c postgres) ConvertValue(v interface{}) (driver.Value, error) {
	v, err := driver.DefaultParameterConverter.ConvertValue(v)
	if err != nil {
		return nil, err
	}
	switch v := v.(type) {
	default:
		return v, nil
	case time.Time:
		return sql.FormatTime(v, sql.TimeLayoutWithOffset), nil
	}
}

func (postgres) errorMapper(err error) error {
	if err == nil {
		return nil
	}

	var (
		msg            = err.Error()
		constraintType = sql.ExtractString(msg, "violates ", " constraint")
	)

	switch constraintType {
	case "unique":
		return dbm.ConstraintError{
			Key:  sql.ExtractString(err.Error(), "constraint \"", "\""),
			Type: dbm.UniqueConstraint,
			Err:  err,
		}
	case "foreign key":
		return dbm.ConstraintError{
			Key:  sql.ExtractString(err.Error(), "constraint \"", "\""),
			Type: dbm.ForeignKeyConstraint,
			Err:  err,
		}
	case "check":
		return dbm.ConstraintError{
			Key:  sql.ExtractString(err.Error(), "constraint \"", "\""),
			Type: dbm.CheckConstraint,
			Err:  err,
		}
	default:
		return err
	}
}

func (postgres) columnMapper(column *dbm.Column) (string, int, int) {
	var (
		typ  string
		m, n int
	)

	// postgres specific
	column.Unsigned = false
	if column.Default == "" {
		column.Default = nil
	}

	switch column.Type {
	case dbm.ID:
		typ = "SERIAL NOT NULL"
	case dbm.BigID:
		typ = "BIGSERIAL NOT NULL"
	case dbm.DateTime:
		typ = "TIMESTAMPTZ"
		if t, ok := column.Default.(time.Time); ok {
			column.Default = sql.FormatTime(t, sql.TimeLayoutWithOffset)
		}
	case dbm.Int, dbm.BigInt, dbm.Text:
		column.Limit = 0
		typ, m, n = sql.ColumnMapper(column)
	case dbm.JSON:
		typ = "JSONB"
	default:
		typ, m, n = sql.ColumnMapper(column)
	}

	return typ, m, n
}
