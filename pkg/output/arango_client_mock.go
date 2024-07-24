package output

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"
)

type ArangoClientMock struct {
	returnDbError bool
	setQueryError bool
	arangoMock    *ArangoDatabaseMock
}

func NewArangoClientMock() *ArangoClientMock {
	return &ArangoClientMock{
		returnDbError: false,
	}
}
func (client *ArangoClientMock) Database(ctx context.Context, name string) (driver.Database, error) {
	if client.returnDbError {
		return nil, fmt.Errorf("error getting DB %s", name)
	}
	if client.setQueryError {
		arangoDbMock := NewArangoDatabaseMock()
		arangoDbMock.returnQueryError = true
		return arangoDbMock, nil
	}
	if client.arangoMock != nil {
		return client.arangoMock, nil
	}
	client.arangoMock = NewArangoDatabaseMock()
	return client.arangoMock, nil
}
