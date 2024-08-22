package output

import (
	"context"

	"github.com/arangodb/go-driver"
)

type ArangoDatabase interface {
	Query(ctx context.Context, query string, bindVars map[string]interface{}) (driver.Cursor, error)
	Collection(ctx context.Context, name string) (driver.Collection, error)
}
