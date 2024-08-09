package output

import (
	"context"

	"github.com/arangodb/go-driver"
)

type ArangoCursor interface {
	ReadDocument(ctx context.Context, result interface{}) (driver.DocumentMeta, error)
	Count() int64
}
