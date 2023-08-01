package adapter

import (
	"github.com/jiyeyuran/dbm/adapter/sql"
	"github.com/jiyeyuran/dbm/adapter/sql/builder"
)

var SQLite3 = func() *sql.SQL {
	var (
		sqlite3          = sqlite3{}
		ddlBufferFactory = builder.BufferFactory{InlineValues: true, BoolTrueValue: "1", BoolFalseValue: "0", Quoter: builder.Quote{IDPrefix: "\"", IDSuffix: "\"", IDSuffixEscapeChar: "\"", ValueQuote: "'", ValueQuoteEscapeChar: "'"}}
		tableBuilder     = builder.Table{BufferFactory: ddlBufferFactory, ColumnMapper: sqlite3.columnMapper, DefinitionFilter: sqlite3.definitionFilter}
		indexBuilder     = builder.Index{BufferFactory: ddlBufferFactory}
	)
	return &sql.SQL{
		TableBuilder: tableBuilder,
		IndexBuilder: indexBuilder,
		ErrorMapper:  sqlite3.errorMapper,
	}
}()

var MYSQL = func() *sql.SQL {
	var (
		mysql            = mysql{}
		ddlBufferFactory = builder.BufferFactory{InlineValues: true, BoolTrueValue: "true", BoolFalseValue: "false", Quoter: mysql, ValueConverter: mysql}
		tableBuilder     = builder.Table{BufferFactory: ddlBufferFactory, ColumnMapper: mysql.columnMapper, DropKeyMapper: mysql.dropKeyMapper}
		indexBuilder     = builder.Index{BufferFactory: ddlBufferFactory, DropIndexOnTable: true}
	)
	return &sql.SQL{
		TableBuilder: tableBuilder,
		IndexBuilder: indexBuilder,
		ErrorMapper:  mysql.errorMapper,
	}
}()

var MSSQL = func() *sql.SQL {
	var (
		mssql            = mssql{}
		ddlBufferFactory = builder.BufferFactory{InlineValues: true, BoolTrueValue: "1", BoolFalseValue: "0", Quoter: builder.Quote{IDPrefix: "[", IDSuffix: "]", IDSuffixEscapeChar: "]", ValueQuote: "'", ValueQuoteEscapeChar: "'"}}
		tableBuilder     = builder.Table{BufferFactory: ddlBufferFactory, ColumnMapper: mssql.columnMapper, DropKeyMapper: sql.DropKeyMapper}
		indexBuilder     = builder.Index{BufferFactory: ddlBufferFactory}
	)

	return &sql.SQL{
		TableBuilder: tableBuilder,
		IndexBuilder: indexBuilder,
		ErrorMapper:  mssql.errorMapper,
	}
}()

var PostgresSQL = func() *sql.SQL {
	var (
		postgres         = postgres{}
		ddlBufferFactory = builder.BufferFactory{InlineValues: true, BoolTrueValue: "true", BoolFalseValue: "false", Quoter: postgres, ValueConverter: postgres}
		tableBuilder     = builder.Table{BufferFactory: ddlBufferFactory, ColumnMapper: postgres.columnMapper, DropKeyMapper: sql.DropKeyMapper}
		indexBuilder     = builder.Index{BufferFactory: ddlBufferFactory}
	)

	return &sql.SQL{
		TableBuilder: tableBuilder,
		IndexBuilder: indexBuilder,
		ErrorMapper:  postgres.errorMapper,
	}
}()

func New(driver string) *sql.SQL {
	switch driver {
	case "mysql":
		return MYSQL
	case "postgres", "pgx":
		return PostgresSQL
	case "sqlite3", "sqlite":
		return SQLite3
	case "mssql":
		return MSSQL
	default:
		return nil
	}
}
