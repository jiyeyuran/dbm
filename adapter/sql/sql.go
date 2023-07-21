package sql

import (
	"github.com/jiyeyuran/dbm"
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

func (s SQL) WrapError(err error) error {
	if s.ErrorMapper == nil || err == nil {
		return err
	}
	return s.ErrorMapper(err)
}
