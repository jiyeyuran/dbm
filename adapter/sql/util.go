package sql

import (
	"strings"
	"time"

	"github.com/jiyeyuran/dbm"
)

// DefaultTimeLayout default time layout.
const DefaultTimeLayout = "2006-01-02 15:04:05"

// TimeLayoutWithZone the time layout with offset.
const TimeLayoutWithOffset = "2006-01-02 15:04:05.999999999-07:00"

func FormatTime(t time.Time, layout string) string {
	return t.Truncate(time.Microsecond).Format(layout)
}

func DropKeyMapper(keyType dbm.KeyType) string {
	return "CONSTRAINT"
}

// ColumnMapper function.
func ColumnMapper(column *dbm.Column) (string, int, int) {
	var (
		typ        string
		m, n       int
		timeLayout = DefaultTimeLayout
	)

	switch column.Type {
	case dbm.ID:
		typ = "INT UNSIGNED AUTO_INCREMENT"
	case dbm.BigID:
		typ = "BIGINT UNSIGNED AUTO_INCREMENT"
	case dbm.Bool:
		typ = "BOOL"
	case dbm.Int:
		typ = "INT"
		m = column.Limit
	case dbm.BigInt:
		typ = "BIGINT"
		m = column.Limit
	case dbm.Float:
		typ = "FLOAT"
		m = column.Precision
	case dbm.Decimal:
		typ = "DECIMAL"
		m = column.Precision
		n = column.Scale
	case dbm.String:
		typ = "VARCHAR"
		m = column.Limit
		if m == 0 {
			m = 255
		}
	case dbm.Text:
		typ = "TEXT"
		m = column.Limit
	case dbm.JSON:
		typ = "TEXT"
	case dbm.Date:
		typ = "DATE"
		timeLayout = "2006-01-02"
	case dbm.DateTime:
		typ = "DATETIME"
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

// ExtractString between two string.
func ExtractString(s, left, right string) string {
	var (
		start = strings.Index(s, left)
		end   = strings.LastIndex(s, right)
	)

	if start < 0 || end < 0 || start+len(left) >= end {
		return s
	}

	return s[start+len(left) : end]
}

func toInt64(i any) int64 {
	var result int64

	switch s := i.(type) {
	case int:
		result = int64(s)
	case int64:
		result = s
	case int32:
		result = int64(s)
	case int16:
		result = int64(s)
	case int8:
		result = int64(s)
	case uint:
		result = int64(s)
	case uint64:
		result = int64(s)
	case uint32:
		result = int64(s)
	case uint16:
		result = int64(s)
	case uint8:
		result = int64(s)
	}

	return result
}
