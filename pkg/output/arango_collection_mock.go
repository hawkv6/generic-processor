package output

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"
)

type ArangoCollectionMock struct {
	wantUpdateErr bool
}

func NewArangoCollectionMock(wantUpdateErr bool) *ArangoCollectionMock {
	return &ArangoCollectionMock{
		wantUpdateErr: wantUpdateErr,
	}
}

func (collection *ArangoCollectionMock) Name() string {
	return ""
}

func (collection *ArangoCollectionMock) Database() driver.Database {
	return nil
}

func (collection *ArangoCollectionMock) Status(ctx context.Context) (driver.CollectionStatus, error) {
	return 0, nil
}

func (collection *ArangoCollectionMock) Count(ctx context.Context) (int64, error) {
	return 1, nil
}

func (collection *ArangoCollectionMock) Statistics(ctx context.Context) (driver.CollectionStatistics, error) {
	return driver.CollectionStatistics{}, nil
}

func (collection *ArangoCollectionMock) Revision(ctx context.Context) (string, error) {
	return "", nil
}

func (collection *ArangoCollectionMock) Checksum(ctx context.Context, withRevisions bool, withData bool) (driver.CollectionChecksum, error) {
	return driver.CollectionChecksum{}, nil
}

func (collection *ArangoCollectionMock) Properties(ctx context.Context) (driver.CollectionProperties, error) {
	return driver.CollectionProperties{}, nil
}

func (collection *ArangoCollectionMock) SetProperties(ctx context.Context, options driver.SetCollectionPropertiesOptions) error {
	return nil
}

func (collection *ArangoCollectionMock) Shards(ctx context.Context, details bool) (driver.CollectionShards, error) {
	return driver.CollectionShards{}, nil
}

func (collection *ArangoCollectionMock) Load(ctx context.Context) error {
	return nil
}

func (collection *ArangoCollectionMock) Unload(ctx context.Context) error {
	return nil
}

func (collection *ArangoCollectionMock) Remove(ctx context.Context) error {
	return nil
}

func (collection *ArangoCollectionMock) Truncate(ctx context.Context) error {
	return nil
}

func (collection *ArangoCollectionMock) Rename(ctx context.Context, newName string) error {
	return nil
}

func (collection *ArangoCollectionMock) CreateDocument(ctx context.Context, document interface{}) (driver.DocumentMeta, error) {
	return driver.DocumentMeta{}, nil
}

func (collection *ArangoCollectionMock) CreateDocuments(ctx context.Context, documents interface{}) (driver.DocumentMetaSlice, driver.ErrorSlice, error) {
	return nil, nil, nil
}

func (collection *ArangoCollectionMock) DocumentExists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

// nolint // This method is not used anyway
func (collection *ArangoCollectionMock) EnsureFullTextIndex(ctx context.Context, fields []string, options *driver.EnsureFullTextIndexOptions) (driver.Index, bool, error) {
	return nil, false, nil
}

func (collection *ArangoCollectionMock) EnsureGeoIndex(ctx context.Context, fields []string, options *driver.EnsureGeoIndexOptions) (driver.Index, bool, error) {
	return nil, false, nil
}

func (collection *ArangoCollectionMock) EnsureHashIndex(ctx context.Context, fields []string, options *driver.EnsureHashIndexOptions) (driver.Index, bool, error) {
	return nil, false, nil
}

func (collection *ArangoCollectionMock) EnsurePersistentIndex(ctx context.Context, fields []string, options *driver.EnsurePersistentIndexOptions) (driver.Index, bool, error) {
	return nil, false, nil
}

func (collection *ArangoCollectionMock) EnsureTTLIndex(ctx context.Context, field string, ttl int, options *driver.EnsureTTLIndexOptions) (driver.Index, bool, error) {
	return nil, false, nil
}

func (collection *ArangoCollectionMock) EnsureInvertedIndex(ctx context.Context, options *driver.InvertedIndexOptions) (driver.Index, bool, error) {
	return nil, false, nil
}

func (collection *ArangoCollectionMock) EnsureSkipListIndex(ctx context.Context, fields []string, options *driver.EnsureSkipListIndexOptions) (driver.Index, bool, error) {
	return nil, false, nil
}

func (collection *ArangoCollectionMock) EnsureZKDIndex(ctx context.Context, fields []string, options *driver.EnsureZKDIndexOptions) (driver.Index, bool, error) {
	return nil, false, nil
}

func (collection *ArangoCollectionMock) ImportDocuments(ctx context.Context, documents interface{}, options *driver.ImportDocumentOptions) (driver.ImportDocumentStatistics, error) {
	return driver.ImportDocumentStatistics{}, nil
}

func (collection *ArangoCollectionMock) Index(ctx context.Context, index string) (driver.Index, error) {
	return nil, nil
}

func (collection *ArangoCollectionMock) Indexes(ctx context.Context) ([]driver.Index, error) {
	return nil, nil
}

func (collection *ArangoCollectionMock) IndexExists(ctx context.Context, index string) (bool, error) {
	return false, nil
}

func (collection *ArangoCollectionMock) ReadDocument(ctx context.Context, key string, result interface{}) (driver.DocumentMeta, error) {
	return driver.DocumentMeta{}, nil
}

func (collection *ArangoCollectionMock) ReadDocuments(ctx context.Context, keys []string, result interface{}) (driver.DocumentMetaSlice, driver.ErrorSlice, error) {
	return nil, nil, nil
}

func (collection *ArangoCollectionMock) RemoveDocument(ctx context.Context, key string) (driver.DocumentMeta, error) {
	return driver.DocumentMeta{}, nil
}

func (collection *ArangoCollectionMock) RemoveDocuments(ctx context.Context, keys []string) (driver.DocumentMetaSlice, driver.ErrorSlice, error) {
	return nil, nil, nil
}

func (collection *ArangoCollectionMock) ReplaceDocument(ctx context.Context, key string, document interface{}) (driver.DocumentMeta, error) {
	return driver.DocumentMeta{}, nil
}

func (collection *ArangoCollectionMock) ReplaceDocuments(ctx context.Context, keys []string, documents interface{}) (driver.DocumentMetaSlice, driver.ErrorSlice, error) {
	return nil, nil, nil
}

func (collection *ArangoCollectionMock) UpdateDocument(ctx context.Context, key string, document interface{}) (driver.DocumentMeta, error) {
	if collection.wantUpdateErr {
		return driver.DocumentMeta{}, fmt.Errorf("error updating document")
	}
	return driver.DocumentMeta{}, nil
}

func (collection *ArangoCollectionMock) UpdateDocuments(ctx context.Context, keys []string, documents interface{}) (driver.DocumentMetaSlice, driver.ErrorSlice, error) {
	if collection.wantUpdateErr {
		return nil, nil, fmt.Errorf("error updating documents")
	}
	return nil, nil, nil
}
