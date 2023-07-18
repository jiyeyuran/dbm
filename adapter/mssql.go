package adapter

import (
	"strings"
	"time"

	"github.com/jiyeuran/dbm"
	"github.com/jiyeuran/dbm/adapter/sql"
)

type mssql struct{}

func (mssql) errorMapper(err error) error {
	if err == nil {
		return nil
	}

	var msg = err.Error()

	switch {
	case strings.HasPrefix(msg, "mssql: Violation of PRIMARY KEY"):
		return dbm.ConstraintError{
			Key:  sql.ExtractString(msg, "constraint '", "'"),
			Type: dbm.UniqueConstraint,
			Err:  err,
		}
	case strings.HasPrefix(msg, "mssql: Violation of UNIQUE KEY"):
		return dbm.ConstraintError{
			Key:  sql.ExtractString(msg, "constraint '", "'"),
			Type: dbm.UniqueConstraint,
			Err:  err,
		}
	case strings.HasPrefix(msg, "mssql: The UPDATE statement conflicted with the FOREIGN KEY"):
		return dbm.ConstraintError{
			Key:  sql.ExtractString(msg, "FOREIGN KEY constraint \"", "\""),
			Type: dbm.ForeignKeyConstraint,
			Err:  err,
		}
	case strings.HasPrefix(msg, "mssql: The UPDATE statement conflicted with the CHECK"):
		return dbm.ConstraintError{
			Key:  sql.ExtractString(msg, "FOREIGN KEY constraint \"", "\""),
			Type: dbm.CheckConstraint,
			Err:  err,
		}
	default:
		return err
	}
}

// columnMapper function.
func (mssql) columnMapper(column *dbm.Column) (string, int, int) {
	var (
		typ        string
		m, n       int
		timeLayout = "2006-01-02 15:04:05"
	)

	switch column.Type {
	case dbm.ID:
		typ = "INT NOT NULL IDENTITY(1,1)"
	case dbm.BigID:
		typ = "BIGINT NOT NULL IDENTITY(1,1)"
	case dbm.Bool:
		typ = "BIT"
	case dbm.Int:
		typ = "INT"
	case dbm.BigInt:
		typ = "BIGINT"
	case dbm.Float:
		typ = "FLOAT"
		m = column.Precision
	case dbm.Decimal:
		typ = "DECIMAL"
		m = column.Precision
		n = column.Scale
	case dbm.String:
		typ = "NVARCHAR"
		m = column.Limit
		if m == 0 {
			m = 255
		} else if m > 4000 {
			m = 4000
		}
	case dbm.Text, dbm.JSON:
		typ = "NVARCHAR(MAX)"
	case dbm.Date:
		typ = "DATE"
		timeLayout = "2006-01-02"
	case dbm.DateTime:
		typ = "DATETIMEOFFSET"
	case dbm.Time:
		typ = "TIME"
		timeLayout = "15:04:05"
	default:
		typ = string(column.Type)
	}

	if t, ok := column.Default.(time.Time); ok {
		column.Default = t.Format(timeLayout)
	}

	return typ, m, n
}
