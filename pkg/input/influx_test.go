package input

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/influxdb/models"
	"github.com/stretchr/testify/assert"
)

func TestNewInfluxInput(t *testing.T) {
	tests := []struct {
		name     string
		wantErr  bool
		url      string
		username string
		password string
		db       string
		timeout  uint
	}{
		{
			name:     "TestNewInfluxInput success",
			url:      "http://localhost:8086",
			username: "admin",
			password: "admin",
			db:       "test",
			timeout:  10,
		},
		{
			name:     "TestNewInfluxInput invalid input",
			url:      "",
			username: "admin",
			password: "admin",
			db:       "test",
			timeout:  0,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.InfluxInputConfig{
				URL:      tt.url,
				Username: tt.username,
				Password: tt.password,
				DB:       tt.db,
				Timeout:  tt.timeout,
			}
			input, err := NewInfluxInput(config, make(chan message.Command), make(chan message.Result))
			if tt.wantErr {
				assert.Nil(t, input)
				assert.Error(t, err)
				return
			}
			assert.NotNil(t, input)
			assert.NoError(t, err)
		})
	}
}

func TestInfluxInput_Init(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		timeout uint
	}{
		{
			name:    "TestInfluxInput_Init ping success",
			timeout: 1,
			wantErr: false,
		},
		{
			name:    "TestInfluxInput_Init ping error",
			timeout: 0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.InfluxInputConfig{
				URL:      "http://localhost:8086",
				Username: "admin",
				Password: "admin",
				DB:       "test",
				Timeout:  tt.timeout,
			}
			input, err := NewInfluxInput(config, make(chan message.Command), make(chan message.Result))
			assert.NoError(t, err)
			input.client = NewInfluxClientMock()
			if tt.wantErr {
				assert.Error(t, input.Init())
				return
			}
			assert.NoError(t, input.Init())
		})
	}
}

func TestInfluxInput_queryDB(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		result  []message.Result
		wantErr bool
	}{
		{
			name:    "TestInfluxInput_queryDB valid query",
			query:   "SELECT * FROM test",
			result:  []message.Result{},
			wantErr: false,
		},
		{
			name:    "TestInfluxInput_queryDB query error",
			query:   "query error",
			result:  nil,
			wantErr: true,
		},
		{
			name:    "TestInfluxInput_queryDB response error",
			query:   "response error",
			result:  nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.InfluxInputConfig{
				URL:      "http://localhost:8086",
				Username: "admin",
				Password: "admin",
				DB:       "test",
				Timeout:  10,
			}
			input, err := NewInfluxInput(config, make(chan message.Command), make(chan message.Result))
			input.client = NewInfluxClientMock()
			assert.NoError(t, err)
			res, err := input.queryDB(tt.query)
			if tt.wantErr {
				assert.Nil(t, res)
				assert.Error(t, err)
				return
			}
			assert.NotNil(t, tt.result)
			assert.NoError(t, err)
		})
	}
}

func TestInflluxInput_createQuery(t *testing.T) {
	tests := []struct {
		name    string
		command message.InfluxQueryCommand
		want    string
	}{
		{
			name: "TestInfluxInput_createQuery no transformation",
			command: message.InfluxQueryCommand{
				Method:      "mean",
				Field:       "value",
				Measurement: "test",
				Interval:    10,
				GroupBy:     []string{"tag1", "tag2"},
			},
			want: "SELECT mean(\"value\") FROM \"test\" WHERE time > now() - 10s AND time <= now() GROUP BY \"tag1\", \"tag2\"",
		},
		{
			name: "TestInfluxInput_createQuery with transformation",
			command: message.InfluxQueryCommand{
				Method:      "mean",
				Field:       "value",
				Measurement: "test",
				Interval:    10,
				GroupBy:     []string{"tag1", "tag2"},
				Transformation: &config.Transformation{
					Operation: "derivative",
					Period:    2,
				},
			},
			want: "SELECT derivative(mean(\"value\"),2s) FROM \"test\" WHERE time > now() - 10s AND time <= now() GROUP BY time(10s),  \"tag1\", \"tag2\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.InfluxInputConfig{
				URL:      "http://localhost:8086",
				Username: "admin",
				Password: "admin",
				DB:       "test",
				Timeout:  10,
			}
			input, err := NewInfluxInput(config, make(chan message.Command), make(chan message.Result))
			assert.NoError(t, err)
			query := input.createQuery(tt.command)
			assert.Equal(t, tt.want, query)
		})
	}
}

func TestInfluxInput_castValue(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    float64
		wantErr bool
	}{
		{
			name:    "TestInfluxInput_castValue int",
			value:   json.Number("1"),
			want:    1,
			wantErr: false,
		},
		{
			name:    "TestInfluxInput_castValue float",
			value:   json.Number("1.0"),
			want:    1.0,
			wantErr: false,
		},
		{
			name:    "TestInfluxInput_castValue string",
			value:   "NaN",
			want:    1.0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.InfluxInputConfig{
				URL:      "http://localhost:8086",
				Username: "admin",
				Password: "admin",
				DB:       "test",
				Timeout:  10,
			}
			input, err := NewInfluxInput(config, make(chan message.Command), make(chan message.Result))
			assert.NoError(t, err)
			value, err := input.castValue(tt.value)
			if tt.wantErr {
				assert.Zero(t, value)
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tt.want, value)
			assert.NoError(t, err)
		})
	}
}

func TestInfluxInput_getResults(t *testing.T) {
	tests := []struct {
		name      string
		command   message.InfluxQueryCommand
		result    []client.Result
		hasValues bool
	}{
		{
			name: "TestInfluxInput_getResults success",
			command: message.InfluxQueryCommand{
				Method:      "mean",
				Field:       "average-delay",
				Measurement: "performance-measurement",
				Interval:    10,
				GroupBy:     []string{"interface_name", "source"},
			},
			result: []client.Result{
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
			hasValues: true,
		},
		{
			name: "TestInfluxInput_getResults conversion error",
			command: message.InfluxQueryCommand{
				Method:      "mean",
				Field:       "average-delay",
				Measurement: "performance-measurement",
				Interval:    10,
				GroupBy:     []string{"interface_name", "source"},
			},
			result: []client.Result{
				{
					Series: []models.Row{
						{
							Name:    "performance-measurement",
							Tags:    map[string]string{"interface_name": "GigabitEthernet0/0/0/0", "source": "SITE-A"},
							Columns: []string{"time", "mean"},
							Values: [][]interface{}{
								{"2021-08-01T00:00:00Z", "1998.5"}, // no json.Number
								{"2021-08-01T00:00:10Z", "1999"},   // no json.Number
								{"2021-08-01T00:00:20Z", "2000"},   // no json.Number
							},
						},
					},
				},
			},
			hasValues: false,
		},
		{
			name: "TestInfluxInput_getResults no values",
			command: message.InfluxQueryCommand{
				Method:      "mean",
				Field:       "average-delay",
				Measurement: "performance-measurement",
				Interval:    10,
				GroupBy:     []string{"interface_name", "source"},
			},
			result: []client.Result{
				{
					Series: []models.Row{
						{
							Name:    "performance-measurement",
							Tags:    map[string]string{"interface_name": "GigabitEthernet0/0/0/0", "source": "SITE-A"},
							Columns: []string{"time", "mean"},
							Values:  [][]interface{}{},
						},
					},
				},
			},
			hasValues: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.InfluxInputConfig{
				URL:      "http://localhost:8086",
				Username: "admin",
				Password: "admin",
				DB:       "test",
				Timeout:  10,
			}
			input, err := NewInfluxInput(config, make(chan message.Command), make(chan message.Result))
			assert.NoError(t, err)
			resultMsg := message.InfluxResultMessage{}
			input.getResults(tt.result, tt.command, &resultMsg)
			if !tt.hasValues {
				assert.Zero(t, resultMsg.Results)
				return
			}
			assert.NotNil(t, resultMsg.Results)
		})
	}
}

func TestInfluxInput_sendResults(t *testing.T) {
	tests := []struct {
		name    string
		results []client.Result
		command message.InfluxQueryCommand
		wantErr bool
	}{
		{
			name: "TestInfluxInput_sendResults success",
			results: []client.Result{
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
			command: message.InfluxQueryCommand{
				Method:      "mean",
				Field:       "average-delay",
				Measurement: "performance-measurement",
				Interval:    10,
				GroupBy:     []string{"interface_name", "source"},
			},
			wantErr: false,
		},
		{
			name:    "TestInfluxInput_sendResults no results",
			results: []client.Result{},
			command: message.InfluxQueryCommand{
				Method:      "mean",
				Field:       "average-delay",
				Measurement: "performance-measurement",
				Interval:    10,
				GroupBy:     []string{"interface_name", "source"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.InfluxInputConfig{
				URL:      "http://localhost:8086",
				Username: "admin",
				Password: "admin",
				DB:       "test",
				Timeout:  10,
			}
			input, err := NewInfluxInput(config, make(chan message.Command), make(chan message.Result))
			assert.NoError(t, err)
			input.client = NewInfluxClientMock()
			if tt.wantErr {
				input.sendResults(tt.results, tt.command)
				return
			}
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				<-input.resultChan
				t.Log("Received result")
				wg.Done()
			}()
			input.sendResults(tt.results, tt.command)
			wg.Wait()
		})
	}
}

func TestInfluxInput_executeCommand(t *testing.T) {
	tests := []struct {
		name    string
		command message.InfluxQueryCommand
		wantErr bool
	}{
		{
			name: "TestInfluxInput_executeCommand success",
			command: message.InfluxQueryCommand{
				Method:      "error",
				Field:       "average-delay",
				Measurement: "performance-measurement",
				Interval:    10,
				GroupBy:     []string{"interface_name", "source"},
			},
			wantErr: true,
		},
		{
			name: "TestInfluxInput_executeCommand empty query",
			command: message.InfluxQueryCommand{
				Method:      "empty",
				Field:       "average-delay",
				Measurement: "performance-measurement",
				Interval:    10,
				GroupBy:     []string{"interface_name", "source"},
			},
			wantErr: true,
		},
		{
			name: "TestInfluxInput_executeCommand success",
			command: message.InfluxQueryCommand{
				Method:      "mean",
				Field:       "average-delay",
				Measurement: "performance-measurement",
				Interval:    10,
				GroupBy:     []string{"interface_name", "source"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.InfluxInputConfig{
				URL:      "http://localhost:8086",
				Username: "admin",
				Password: "admin",
				DB:       "test",
				Timeout:  10,
			}
			input, err := NewInfluxInput(config, make(chan message.Command), make(chan message.Result))
			assert.NoError(t, err)
			input.client = NewInfluxClientMock()
			if tt.wantErr {
				input.executeCommand(tt.command)
				return
			}
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				<-input.resultChan
				t.Log("Received result")
				wg.Done()
			}()
			input.executeCommand(tt.command)
			wg.Wait()
		})
	}
}

func TestInfluxInput_Start(t *testing.T) {
	tests := []struct {
		name    string
		command message.Command
		wantErr bool
	}{
		{
			name:    "TestInfluxInput_Start wrong command",
			command: message.ArangoUpdateCommand{},
			wantErr: true,
		},
		{
			name: "TestInfluxInput_Start success",
			command: message.InfluxQueryCommand{
				Method:      "mean",
				Field:       "average-delay",
				Measurement: "performance-measurement",
				Interval:    10,
				GroupBy:     []string{"interface_name", "source"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.InfluxInputConfig{
				URL:      "http://localhost:8080",
				Username: "admin",
				Password: "admin",
				DB:       "test",
				Timeout:  10,
			}
			input, err := NewInfluxInput(config, make(chan message.Command), make(chan message.Result))
			assert.NoError(t, err)
			input.client = NewInfluxClientMock()
			if tt.wantErr {
				go func() {
					input.Start()
					t.Log("Received result")
				}()
				time.Sleep(100 * time.Millisecond)
				input.commandChan <- tt.command
				close(input.quitChan)
				input.wg.Wait()
				return
			}
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				<-input.resultChan
				t.Log("Received result")
				wg.Done()
			}()
			go func() {
				input.Start()
			}()
			input.commandChan <- tt.command
			wg.Wait()
			close(input.quitChan)
			input.wg.Wait()
		})
	}
}

func TestInfluxInput_Stop(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestInfluxInput_Stop success",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.InfluxInputConfig{
				URL:      "http://localhost:8080",
				Username: "admin",
				Password: "admin",
				DB:       "test",
				Timeout:  10,
			}
			input, err := NewInfluxInput(config, make(chan message.Command), make(chan message.Result))
			input.client = NewInfluxClientMock()
			assert.NoError(t, err)
			go func() {
				input.Start()
			}()
			time.Sleep(100 * time.Millisecond)
			assert.NoError(t, input.Stop())
			input.wg.Wait()
		})
	}
}
