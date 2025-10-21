package postgres

import (
	"fmt"

	"github.com/wb-go/wbf/dbpg"
)

const (
	tableLinks = "links"
)

// NewMSPostgresDB создает новое подключение к базе данных postgres, поддерживающее
// master-slave масштабирование.
func NewMSPostgresDB(host, port, username, dbName, password, sslMode string) (*dbpg.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		host, port, username, dbName, password, sslMode)

	return dbpg.New(dsn, []string{}, &dbpg.Options{})
}
