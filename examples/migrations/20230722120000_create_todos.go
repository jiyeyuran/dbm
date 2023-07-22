package migrations

import (
	"context"
	"fmt"
	"time"

	"github.com/jiyeyuran/dbm"
)

// MigrateCreateTodos definition
func MigrateCreateTodos(schema *dbm.Schema) {
	schema.CreateTable("todos", func(t *dbm.Table) {
		t.ID("id")
		t.DateTime("created_at")
		t.DateTime("updated_at")
		t.String("title", dbm.Required(true), dbm.Limit(255))
		t.Bool("completed", dbm.Required(true), dbm.Default(0))
		t.Int("order", dbm.Required(true), dbm.Default(0))
	})

	schema.CreateIndex("todos", "order", []string{"order"})

	schema.Do(func(ctx context.Context, db dbm.Database) error {
		// add seeds
		now := time.Now().Format("2006-01-02 15:04:05")
		sqlstr := fmt.Sprintf(
			"insert into todos(created_at,updated_at,title) values('%s','%s','%s')",
			now, now, "Do Homework",
		)
		// db is the database instance passed in dbm.New.
		_, err := db.ExecContext(ctx, sqlstr)
		return err
	})
}

// RollbackCreateTodos definition
func RollbackCreateTodos(schema *dbm.Schema) {
	schema.DropTable("todos")
}
