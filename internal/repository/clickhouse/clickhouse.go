package clickhouse

import (
	"database/sql"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

const (
	tableTransits = "transits"
)

// NewClickhouseDB создает новое подключение к базе данных Clickhouse.
func NewClickhouseDB(url, username, password, database string) (*sql.DB, error) {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{url},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: time.Second * 30,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Debug:                false,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10240,
	})

	return conn, conn.Ping()
}
