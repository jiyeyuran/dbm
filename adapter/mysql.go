package adapter

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/jiyeuran/dbm"
	"github.com/jiyeuran/dbm/adapter/sql"
)

type mysql struct{}

func (q mysql) ID(name string) string {
	end := strings.IndexRune(name, 0)
	if end > -1 {
		name = name[:end]
	}
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func (q mysql) Value(v interface{}) string {
	switch v := v.(type) {
	default:
		panic("unsupported value")
	case string:
		// TODO: Need to check on connection for NO_BACKSLASH_ESCAPES
		rv := []rune(v)
		buf := make([]rune, len(rv)*2)
		pos := 0
		for i := 0; i < len(rv); i++ {
			c := rv[i]
			switch c {
			case '\x00':
				buf[pos] = '\\'
				buf[pos+1] = '0'
				pos += 2
			case '\n':
				buf[pos] = '\\'
				buf[pos+1] = 'n'
				pos += 2
			case '\r':
				buf[pos] = '\\'
				buf[pos+1] = 'r'
				pos += 2
			case '\x1a':
				buf[pos] = '\\'
				buf[pos+1] = 'Z'
				pos += 2
			case '\'':
				buf[pos] = '\\'
				buf[pos+1] = '\''
				pos += 2
			case '"':
				buf[pos] = '\\'
				buf[pos+1] = '"'
				pos += 2
			case '\\':
				buf[pos] = '\\'
				buf[pos+1] = '\\'
				pos += 2
			default:
				buf[pos] = c
				pos++
			}
		}

		return "'" + string(buf[:pos]) + "'"
	}
}

func (c mysql) ConvertValue(v interface{}) (driver.Value, error) {
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

func (mysql) errorMapper(err error) error {
	if err == nil {
		return nil
	}

	var (
		msg          = err.Error()
		errCodeSep   = ':'
		errCodeIndex = strings.IndexRune(msg, errCodeSep)
	)

	if errCodeIndex < 0 {
		errCodeIndex = 0
	}

	switch msg[:errCodeIndex] {
	case "Error 1062":
		return dbm.ConstraintError{
			Key:  sql.ExtractString(msg, "key '", "'"),
			Type: dbm.UniqueConstraint,
			Err:  err,
		}
	case "Error 1452":
		return dbm.ConstraintError{
			Key:  sql.ExtractString(msg, "CONSTRAINT `", "`"),
			Type: dbm.ForeignKeyConstraint,
			Err:  err,
		}
	default:
		return err
	}
}

func (mysql) columnMapper(column *dbm.Column) (string, int, int) {
	switch column.Type {
	case dbm.JSON:
		return "JSON", 0, 0

	case dbm.DateTime:
		return "DATETIME", column.Precision, 0

	default:
		return sql.ColumnMapper(column)
	}
}

func (mysql) dropKeyMapper(typ dbm.KeyType) string {
	if typ == dbm.ForeignKey {
		return "FOREIGN KEY"
	}

	panic(fmt.Sprintf("drop key: unsupported key type `%s`", typ))
}
