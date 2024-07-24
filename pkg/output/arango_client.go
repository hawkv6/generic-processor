package output

import (
	"context"

	"github.com/arangodb/go-driver"
)

type ArangoClient interface {
	// Connection returns the connection used by this client

	// ClientDatabases - Database functions
	// driver.ClientDatabases
	Database(ctx context.Context, name string) (driver.Database, error)

	// // ClientUsers - User functions
	// driver.ClientUsers

	// // ClientCluster - Cluster functions
	// ClientCluster

	// // ClientServerInfo - Individual server information functions
	// ClientServerInfo

	// // ClientServerAdmin - Server/cluster administration functions
	// ClientServerAdmin

	// // ClientReplication - Replication functions
	// ClientReplication

	// // ClientAdminBackup - Backup functions
	// ClientAdminBackup

	// ClientFoxx - Foxx functions
	// ClientFoxx

	// ClientAsyncJob - Asynchronous job functions
	// ClientAsyncJob

	// ClientLog
}
