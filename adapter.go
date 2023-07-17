package dbm

import (
	"context"
)

type Adapter interface {
	SchemaApply(ctx context.Context, migration IMigration) error
}
