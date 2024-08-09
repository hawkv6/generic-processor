package output

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"
	"github.com/hawkv6/generic-processor/pkg/message"
)

type ArangoCursorMock struct {
	returnError      bool
	countReturnValue int64
	internalCount    int
}

func NewArangoCursorMock() *ArangoCursorMock {
	return &ArangoCursorMock{
		returnError:      false,
		countReturnValue: 0,
		internalCount:    0,
	}
}

func (cursor *ArangoCursorMock) ReadDocument(ctx context.Context, result interface{}) (driver.DocumentMeta, error) {
	if cursor.returnError {
		return driver.DocumentMeta{}, fmt.Errorf("error reading document")
	}
	if cursor.internalCount > 0 {
		return driver.DocumentMeta{}, driver.NoMoreDocumentsError{}
	}
	localLinkIp := "2001:db8::1"
	cast := result.(*message.UpdateLinkMessage)
	cast.Key = localLinkIp
	cast.LocalLinkIP = localLinkIp
	cast.UnidirLinkDelay = 10
	cast.UnidirLinkDelayMinMax = []uint32{10, 10}
	cast.MaxLinkBWKbps = 10
	cast.UnidirDelayVariation = 10
	cast.UnidirResidualBW = 10
	cast.UnidirAvailableBW = 10
	cast.UnidirBWUtilization = 10
	cast.NormalizedUnidirLinkDelay = 10
	cast.NormalizedUnidirDelayVariation = 10
	cast.NormalizedUnidirPacketLoss = 10
	cast.UnidirPacketLossPercentage = 10

	cursor.internalCount++

	return driver.DocumentMeta{}, nil
}
func (cursor *ArangoCursorMock) Count() int64 {
	return cursor.countReturnValue
}

func (cursor *ArangoCursorMock) Close() error {
	return nil
}

func (cursor *ArangoCursorMock) HasMore() bool {
	return cursor.internalCount == 0
}

func ReadDocument(ctx context.Context, result interface{}) (driver.DocumentMeta, error) {
	return driver.DocumentMeta{}, nil
}

func (cursor *ArangoCursorMock) RetryReadDocument(ctx context.Context, result interface{}) (driver.DocumentMeta, error) {
	return driver.DocumentMeta{}, nil
}

func (cursor *ArangoCursorMock) Statistics() driver.QueryStatistics {
	return nil
}

func (cursor *ArangoCursorMock) Extra() driver.QueryExtra {
	return nil
}
