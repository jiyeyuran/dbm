# dbm

Migration is a feature that allows you to evolve your database schema over time, dbm provides DSL that allows you to write migration in Golang.

# Defining Migration

Migration package usually located inside your-repo/db/migrations package. It's a standalone package that should not be imported by the rest of your application. Each migration file is named as number_name.go, and each migration file must define a pair of migration and rollback functions: MigrateName and RollbackName. Migrate and rollback function name is the camel cased file name without version.

```go
// 20202806225100_create_todos.go

package migrations

import (
    "context"

    "github.com/jiyeyuran/dbm"
)

// MigrateCreateTodos definition
func MigrateCreateTodos(schema *dbm.Schema) {
    schema.CreateTable("todos", func(t *dbm.Table) {
        t.ID("id")
        t.DateTime("created_at")
        t.DateTime("updated_at")
        t.String("title")
        t.Bool("completed")
        t.Int("order")
    })

    schema.CreateIndex("todos", "order", []string{"order"})

    schema.Do(func(ctx context.Context, db dbm.Database) error {
        // TODO: add seeds
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
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    ctx := context.TODO()

    conn, err := sql.Open("mysql", "root@(source:3306)/dbm_test?charset=utf8&parseTime=True&loc=Local")
    check(err)

    m := dbm.New(adapter.MYSQL, conn)

    // Register migrations
    m.Register(20202806225100, migrations.MigrateCreateTodos, migrations.RollbackCreateTodos)

    // Run migrations
    check(m.Migrate(ctx))
    // OR:
    // check(m.Rollback(ctx))
}

func check(err error) {
    if err != nil {
        panic(err)
    }
}
```