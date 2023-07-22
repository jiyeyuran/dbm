# dbm

Migration is a feature that allows you to evolve your database schema over time. dbm, a zero-dependency library, provides DSL that allows you to write migration in Golang.

# Defining Migration

Migration package usually located inside your-repo/db/migrations package. It's a standalone package that should not be imported by the rest of your application. Each migration file is named as number_name.go, and each migration file must define a pair of migration and rollback functions: MigrateName and RollbackName. Migrate and rollback function name is the camel cased file name without version.

```go
// 20230722120000_create_todos.go

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
```

# Run Migrations

```go
package main

import (
    "context"

    "github.com/jiyeyuran/dbm"
    "github.com/jiyeyuran/dbm/adapter"
    "github.com/jiyeyuran/dbm/examples/migrations"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    ctx := context.TODO()

    conn, err := sql.Open("mysql", "root@(localhost:3306)/dbm_test?charset=utf8&parseTime=True&loc=Local")
    check(err)

    m := dbm.New(adapter.MYSQL, conn)

    // It is recommended to use transactions
    // tx, _ := conn.BeginTx(ctx, nil)
    // m := dbm.New(adapter.MYSQL, tx)

    // Register migrations
    m.Register(20230722120000, migrations.MigrateCreateTodos, migrations.RollbackCreateTodos)

    // Run migrations
    check(m.Migrate(ctx))
    // OR:
    // check(m.Rollback(ctx))

    // if using a transaction, don't forget to commit it.
    // tx.Commit()
}

func check(err error) {
    if err != nil {
        panic(err)
    }
}
```