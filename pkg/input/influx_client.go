package input

import (
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

type InfluxClient interface {
	Ping(timeout time.Duration) (time.Duration, string, error)

	Query(q client.Query) (*client.Response, error)

	Close() error
}
