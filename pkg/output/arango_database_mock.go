package output

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"
)

type ArangoDatabaseMock struct {
	returnQueryError           bool
	returnCollectionError      bool
	returnUpdateDocumentsError bool
	cursor                     *ArangoCursorMock
}

func NewArangoDatabaseMock() *ArangoDatabaseMock {
	return &ArangoDatabaseMock{
		returnQueryError:           false,
		returnCollectionError:      false,
		returnUpdateDocumentsError: false,
	}
}

func (db *ArangoDatabaseMock) Query(ctx context.Context, query string, bindVars map[string]interface{}) (driver.Cursor, error) {
	if db.returnQueryError {
		return nil, fmt.Errorf("error querying DB")
	}
	if db.cursor != nil {
		return db.cursor, nil
	}
	db.cursor = NewArangoCursorMock()
	return db.cursor, nil
}

func (db *ArangoDatabaseMock) Collection(ctx context.Context, name string) (driver.Collection, error) {
	if db.returnCollectionError {
		return nil, fmt.Errorf("error getting collection %s", name)
	} else {
		return NewArangoCollectionMock(db.returnUpdateDocumentsError), nil

	}
}

func (db *ArangoDatabaseMock) Name() string {
	return ""
}

func (db *ArangoDatabaseMock) Info(ctx context.Context) (driver.DatabaseInfo, error) {
	return driver.DatabaseInfo{}, nil
}

func (db *ArangoDatabaseMock) EngineInfo(ctx context.Context) (driver.EngineInfo, error) {
	return driver.EngineInfo{}, nil
}

func (db *ArangoDatabaseMock) Remove(ctx context.Context) error {
	return nil
}

func (db *ArangoDatabaseMock) ValidateQuery(ctx context.Context, query string) error {
	return nil
}

func (db *ArangoDatabaseMock) ExplainQuery(ctx context.Context, query string, bindVars map[string]interface{}, opts *driver.ExplainQueryOptions) (driver.ExplainQueryResult, error) {
	return driver.ExplainQueryResult{}, nil
}

func (db *ArangoDatabaseMock) OptimizerRulesForQueries(ctx context.Context) ([]driver.QueryRule, error) {
	return nil, nil
}

func (db *ArangoDatabaseMock) Transaction(ctx context.Context, action string, opts *driver.TransactionOptions) (interface{}, error) {
	return driver.TransactionCollections{}, nil
}

func (db *ArangoDatabaseMock) AbortTransaction(ctx context.Context, transaction driver.TransactionID, opts *driver.AbortTransactionOptions) error {
	return nil
}

func (db *ArangoDatabaseMock) Analyzer(ctx context.Context, name string) (driver.ArangoSearchAnalyzer, error) {
	return nil, nil
}

func (db *ArangoDatabaseMock) Analyzers(ctx context.Context) ([]driver.ArangoSearchAnalyzer, error) {
	return nil, nil
}

func (db *ArangoDatabaseMock) BeginTransaction(ctx context.Context, transcactions driver.TransactionCollections, opts *driver.BeginTransactionOptions) (driver.TransactionID, error) {
	return "", nil
}

func (db *ArangoDatabaseMock) CancelJob(ctx context.Context, id string) error {
	return nil
}

func (db *ArangoDatabaseMock) CollectionExists(ctx context.Context, name string) (bool, error) {
	return false, nil
}

func (db *ArangoDatabaseMock) Collections(ctx context.Context) ([]driver.Collection, error) {
	return nil, nil
}

func (db *ArangoDatabaseMock) CommitTransaction(ctx context.Context, transaction driver.TransactionID, opts *driver.CommitTransactionOptions) error {
	return nil
}

func (db *ArangoDatabaseMock) CreateArangoSearchAliasView(ctx context.Context, name string, options *driver.ArangoSearchAliasViewProperties) (driver.ArangoSearchViewAlias, error) {
	return nil, nil
}

func (db *ArangoDatabaseMock) CreateArangoSearchView(ctx context.Context, name string, options *driver.ArangoSearchViewProperties) (driver.ArangoSearchView, error) {
	return nil, nil
}

func (db *ArangoDatabaseMock) CreateCollection(ctx context.Context, name string, opts *driver.CreateCollectionOptions) (driver.Collection, error) {
	return nil, nil
}

func (db *ArangoDatabaseMock) CreateGraph(ctx context.Context, name string, opts *driver.CreateGraphOptions) (driver.Graph, error) {
	return nil, nil
}

func (db *ArangoDatabaseMock) CreateGraphV2(ctx context.Context, name string, opts *driver.CreateGraphOptions) (driver.Graph, error) {
	return nil, nil
}

func (db *ArangoDatabaseMock) EnsureAnalyzer(ctx context.Context, definition driver.ArangoSearchAnalyzerDefinition) (bool, driver.ArangoSearchAnalyzer, error) {
	return false, nil, nil
}

func (db *ArangoDatabaseMock) GetJob(ctx context.Context, id string) (*driver.PregelJob, error) {
	return &driver.PregelJob{}, nil
}

func (db *ArangoDatabaseMock) GetJobs(ctx context.Context) ([]*driver.PregelJob, error) {
	return nil, nil
}

func (db *ArangoDatabaseMock) Graph(ctx context.Context, name string) (driver.Graph, error) {
	return nil, nil
}

func (db *ArangoDatabaseMock) GraphExists(ctx context.Context, name string) (bool, error) {
	return false, nil
}

func (db *ArangoDatabaseMock) Graphs(ctx context.Context) ([]driver.Graph, error) {
	return nil, nil
}

func (db *ArangoDatabaseMock) StartJob(ctx context.Context, job driver.PregelJobOptions) (string, error) {
	return "", nil
}

func (db *ArangoDatabaseMock) TransactionStatus(ctx context.Context, id driver.TransactionID) (driver.TransactionStatusRecord, error) {
	return driver.TransactionStatusRecord{}, nil
}

func (db *ArangoDatabaseMock) View(ctx context.Context, name string) (driver.View, error) {
	return nil, nil
}

func (db *ArangoDatabaseMock) ViewExists(ctx context.Context, name string) (bool, error) {
	return false, nil
}

func (db *ArangoDatabaseMock) Views(ctx context.Context) ([]driver.View, error) {
	return nil, nil
}
