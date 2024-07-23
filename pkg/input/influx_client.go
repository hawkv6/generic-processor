package input

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/influxdb/models"
)

type InfluxClient interface {
	Ping(timeout time.Duration) (time.Duration, string, error)

	Query(q client.Query) (*client.Response, error)

	Close() error
}

type InfluxClientMock struct {
}

func NewInfluxClientMock() *InfluxClientMock {
	return &InfluxClientMock{}
}

func (influxMock *InfluxClientMock) Ping(timeout time.Duration) (time.Duration, string, error) {
	if timeout == 0 {
		return 0, "mock", fmt.Errorf("Ping error")
	} else {
		return 0, "mock", nil
	}
}

func (influxMock *InfluxClientMock) Query(q client.Query) (*client.Response, error) {
	errorQuery := "SELECT error(\"average-delay\") FROM \"performance-measurement\" WHERE time > now() - 10s AND time <= now() GROUP BY \"interface_name\", \"source\""
	if q.Command == "query error" || q.Command == errorQuery {
		return nil, fmt.Errorf("Query error")
	}
	if q.Command == "response error" {
		return &client.Response{
			Results: nil,
			Err:     ("Response error"),
		}, nil

	}
	validQuery := "SELECT mean(\"average-delay\") FROM \"performance-measurement\" WHERE time > now() - 10s AND time <= now() GROUP BY \"interface_name\", \"source\""
	if q.Command == validQuery {
		return &client.Response{
			Results: []client.Result{
				{
					Series: []models.Row{
						{
							Name:    "performance-measurement",
							Tags:    map[string]string{"interface_name": "GigabitEthernet0/0/0/0", "source": "SITE-A"},
							Columns: []string{"time", "mean"},
							Values: [][]interface{}{
								{"2021-08-01T00:00:00Z", json.Number("1998.5")},
								{"2021-08-01T00:00:10Z", json.Number("1999")},
								{"2021-08-01T00:00:20Z", json.Number("2000")},
							},
						},
					},
				},
			},
		}, nil

	}
	emptyQuery := "SELECT empty(\"average-delay\") FROM \"performance-measurement\" WHERE time > now() - 10s AND time <= now() GROUP BY \"interface_name\", \"source\""
	if q.Command == emptyQuery {
		return &client.Response{
			Results: []client.Result{},
		}, nil
	} else {
		return &client.Response{
			Results: []client.Result{},
		}, nil
	}
}

func (influxMock *InfluxClientMock) Close() error {
	return nil
}
