package sql

import (
	"context"
	"database/sql"

	"github.com/jiyeuran/dbm"
)

// ErrorMapper function.
type ErrorMapper func(error) error

type SQL struct {
	*sql.DB
	TableBuilder TableBuilder
	IndexBuilder IndexBuilder
	ErrorMapper  ErrorMapper
}

// SchemaApply performs migration to database.
func (s SQL) SchemaApply(ctx context.Context, migration dbm.IMigration) error {
	var statement string

	switch v := migration.(type) {
	case dbm.Table:
		statement = s.TableBuilder.Build(v)
	case dbm.Index:
		statement = s.IndexBuilder.Build(v)
	case dbm.Raw:
		statement = string(v)
	}

	_, err := s.ExecContext(ctx, statement)
	if err != nil {
		return s.ErrorMapper(err)
	}
	return nil
}
