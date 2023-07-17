package dbm

import (
	"context"
	"fmt"
	"sort"
	"time"
)

const versionTable = "dbm_schema_versions"

type version struct {
	ID        int
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time

	up      Schema
	down    Schema
	applied bool
}

func (version) Table() string {
	return versionTable
}

type versions []version

func (v versions) Len() int {
	return len(v)
}

func (v versions) Less(i, j int) bool {
	return v[i].Version < v[j].Version
}

func (v versions) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

// Migration utility that handles migration logic.
type Migration struct {
	db                 Database
	adapter            Adapter
	versions           versions
	versionTableExists bool
}

// Register a migration.
func (m *Migration) Register(v int, up func(schema *Schema), down func(schema *Schema)) {
	var upSchema, downSchema Schema

	up(&upSchema)
	down(&downSchema)

	m.versions = append(m.versions, version{Version: v, up: upSchema, down: downSchema})
}

func (m Migration) buildVersionTableDefinition() Table {
	var schema Schema
	schema.CreateTableIfNotExists(versionTable, func(t *Table) {
		t.ID("id")
		t.BigInt("version", Unsigned(true), Unique(true))
		t.DateTime("created_at")
		t.DateTime("updated_at")
	})

	return schema.Migrations[0].(Table)
}

func (m *Migration) sync(ctx context.Context) {
	var (
		versions versions
		vi       int
	)

	if !m.versionTableExists {
		check(m.adapter.SchemaApply(ctx, m.buildVersionTableDefinition()))
		m.versionTableExists = true
	}
	sqlstr := "SELECT id, version, created_at, updated_at FROM " + versionTable + " ORDER BY version"
	rows, err := m.db.QueryContext(ctx, sqlstr)
	check(err)
	defer rows.Close()

	for rows.Next() {
		ver := version{}
		err = rows.Scan(&ver.ID, &ver.Version, &ver.CreatedAt, &ver.UpdatedAt)
		check(err)
		versions = append(versions, ver)
	}

	sort.Sort(m.versions)

	for i := range m.versions {
		if vi < len(versions) && m.versions[i].Version == versions[vi].Version {
			m.versions[i].ID = versions[vi].ID
			m.versions[i].applied = true
			vi++
		} else {
			m.versions[i].applied = false
		}
	}

	if vi != len(versions) {
		panic(fmt.Sprint("dbm: missing local migration: ", versions[vi].Version))
	}
}

// Migrate to the latest schema version.
func (m *Migration) Migrate(ctx context.Context) {
	m.sync(ctx)

	for _, v := range m.versions {
		if v.applied {
			continue
		}

		sqlstr := fmt.Sprintf("INSERT INTO %s(version, created_at, updated_at) VALUES (%d, %q, %q)",
			versionTable, v.Version, v.CreatedAt, v.UpdatedAt)
		_, err := m.db.ExecContext(ctx, sqlstr)
		check(err)

		m.run(ctx, v.up.Migrations)
	}
}

// Rollback migration 1 step.
func (m *Migration) Rollback(ctx context.Context) {
	m.sync(ctx)

	for i := range m.versions {
		v := m.versions[len(m.versions)-i-1]
		if !v.applied {
			continue
		}

		sqlstr := fmt.Sprintf("DELETE FROM %s WHERE version=%d", versionTable, v.Version)
		_, err := m.db.ExecContext(ctx, sqlstr)
		check(err)

		m.run(ctx, v.down.Migrations)

		// only rollback one version.
		return
	}
}

func (m *Migration) run(ctx context.Context, migrations []IMigration) {
	for _, migration := range migrations {
		if fn, ok := migration.(Do); ok {
			check(fn(ctx, m.adapter))
		} else {
			check(m.adapter.SchemaApply(ctx, migration))
		}
	}
}

// New migration manager.
func New(adapter Adapter, db Database) Migration {
	return Migration{
		db:      db,
		adapter: adapter,
	}
}

// New migration manager by driver.
func NewByDriver(driver string, db Database) Migration {
	var adapter Adapter

	switch driver {
	case "mysql":

	case "postgres":

	case "sqlserver":

	case "sqlite", "sqlite3":

	}

	return New(adapter, db)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
