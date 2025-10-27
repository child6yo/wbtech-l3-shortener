package postgres

import (
	"github.com/wb-go/wbf/dbpg"
)

const (
	tableLinks = "links"
)

// NewMSPostgresDB создает новое подключение к базе данных postgres, поддерживающее
// master-slave масштабирование.
func NewMSPostgresDB(masterDSN string, slavesDSNs ...string) (*dbpg.DB, error) {
	return dbpg.New(masterDSN, slavesDSNs, &dbpg.Options{})
}
