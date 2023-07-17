package sql

import "github.com/jiyeuran/dbm"

type TableBuilder interface {
	Build(table dbm.Table) string
}

type IndexBuilder interface {
	Build(index dbm.Index) string
}
