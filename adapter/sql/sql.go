package sql

import (
	"github.com/jiyeuran/dbm"
)

// ErrorMapper function.
type ErrorMapper func(error) error

type SQL struct {
	TableBuilder TableBuilder
	IndexBuilder IndexBuilder
	ErrorMapper  ErrorMapper
}

func (s SQL) Build(migration interface{}) string {
	switch v := migration.(type) {
	case dbm.Table:
		return s.TableBuilder.Build(v)
	case dbm.Index:
		return s.IndexBuilder.Build(v)
	case dbm.Raw:
		return string(v)
	}

	return ""
}

func (s SQL) MapError(e error) error {
	return s.ErrorMapper(e)
}
