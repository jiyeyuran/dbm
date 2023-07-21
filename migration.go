package dbm

import (
	"context"
	"fmt"
	"sort"
	"time"
)

const (
	versionTable = "dbm_schema_versions"
	timeLayout   = "2006-01-02 15:04:05-07:00"
)

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

func (m *Migration) sync(ctx context.Context) error {
	var (
		versions versions
		vi       int
	)

	if !m.versionTableExists {
		if err := m.run(ctx, m.buildVersionTableDefinition()); err != nil {
			return err
		}
		m.versionTableExists = true
	}
	sqlstr := "SELECT id, version, created_at, updated_at FROM " + versionTable + " ORDER BY version"
	rows, err := m.db.QueryContext(ctx, sqlstr)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		ver := version{}
		if err = rows.Scan(&ver.ID, &ver.Version, &ver.CreatedAt, &ver.UpdatedAt); err != nil {
			return err
		}
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
		return fmt.Errorf("dbm: missing local migration: %d", versions[vi].Version)
	}
	return nil
}

// Migrate to the latest schema version.
func (m *Migration) Migrate(ctx context.Context) error {
	if err := m.sync(ctx); err != nil {
		return err
	}

	for _, v := range m.versions {
		if v.applied {
			continue
		}
		now := time.Now().Truncate(time.Microsecond).Format(timeLayout)
		sqlstr := fmt.Sprintf("INSERT INTO %s(version, created_at, updated_at) VALUES (%d, %q, %q)",
			versionTable, v.Version, now, now)
		if _, err := m.db.ExecContext(ctx, sqlstr); err != nil {
			return err
		}
		if err := m.run(ctx, v.up.Migrations...); err != nil {
			return err
		}
	}
	return nil
}

// Rollback migration 1 step.
func (m *Migration) Rollback(ctx context.Context) error {
	if err := m.sync(ctx); err != nil {
		return err
	}

	for i := range m.versions {
		v := m.versions[len(m.versions)-i-1]
		if !v.applied {
			continue
		}
		sqlstr := fmt.Sprintf("DELETE FROM %s WHERE version=%d", versionTable, v.Version)
		if _, err := m.db.ExecContext(ctx, sqlstr); err != nil {
			return err
		}
		err := m.run(ctx, v.down.Migrations...)
		// only rollback one version.
		return err
	}
	return nil
}

func (m *Migration) run(ctx context.Context, migrations ...Migratable) error {
	for _, migration := range migrations {
		if fn, ok := migration.(Do); ok {
			if err := fn(ctx, m.db); err != nil {
				return err
			}
		} else {
			if _, err := m.db.ExecContext(ctx, m.adapter.Build(migration)); err != nil {
				if v, ok := m.adapter.(interface{ WrapError(error) error }); ok {
					return v.WrapError(err)
				}
				return err
			}
		}
	}
	return nil
}

// New migration manager.
func New(adapter Adapter, db Database) Migration {
	return Migration{
		db:      db,
		adapter: adapter,
	}
}
