package sql

import "github.com/jiyeyuran/dbm"

type TableBuilder interface {
	Build(table dbm.Table) string
}

type IndexBuilder interface {
	Build(index dbm.Index) string
}
