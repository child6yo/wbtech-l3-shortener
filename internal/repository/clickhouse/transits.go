package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/child6yo/wbtech-l3-shortener/internal/models"
)

// TransitsRepository отвечает за область хранения переходов по ссылкам в Clickhouse.
type TransitsRepository struct {
	db *sql.DB
}

// NewTransitsRepository создает новый TransitsRepository.
func NewTransitsRepository(db *sql.DB) *TransitsRepository {
	return &TransitsRepository{db: db}
}

// CreateTransit добавляет новый переход.
func (tr *TransitsRepository) CreateTransit(ctx context.Context, transit models.Transit) error {
	query := fmt.Sprintf("INSERT INTO %s VALUES ($1, $2, $3)", tableTransits)
	_, err := tr.db.ExecContext(ctx, query, transit.Link, transit.Timestamp, transit.UserAgent)
	if err != nil {
		return fmt.Errorf("repository: %v", err)
	}

	return nil
}

// GetAggregatedTransits количество переходов, агрегированных по query.
// Если в query не прописаны агрегации - агрегирует по user_agent.
func (tr *TransitsRepository) GetAggregatedTransits(
	ctx context.Context,
	query models.TransitAggregationQuery,
) ([]models.TransitAggregationResult, error) {
	var selectFields []string
	var groupByFields []string

	if query.GroupByDay && query.GroupByMonth {
		return nil, fmt.Errorf("repository: cannot group by both day and month")
	}

	selectFields = append(selectFields, "count(*) AS count")

	if query.GroupByDay {
		selectFields = append(selectFields, "toStartOfDay(timestamp) AS date")
		groupByFields = append(groupByFields, "toStartOfDay(timestamp)")
	}
	if query.GroupByMonth {
		selectFields = append(selectFields, "toStartOfMonth(timestamp) AS date")
		groupByFields = append(groupByFields, "toStartOfMonth(timestamp)")
	}
	if query.GroupByUserAgent || len(selectFields) == 1 {
		selectFields = append(selectFields, "user_agent")
		groupByFields = append(groupByFields, "user_agent")
	}

	sqlQuery := fmt.Sprintf(
		"SELECT %s FROM %s WHERE link = ? GROUP BY %s",
		strings.Join(selectFields, ", "),
		tableTransits,
		strings.Join(groupByFields, ", "),
	)

	rows, err := tr.db.QueryContext(ctx, sqlQuery, string(query.Link))

	if err != nil {
		return nil, fmt.Errorf("repository: failed to execute query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("repository: %v", err)
	}

	var results []models.TransitAggregationResult

	for rows.Next() {
		var count uint64
		var date *time.Time
		var userAgent *string

		dest := make([]interface{}, len(columns))
		for i, col := range columns {
			switch col {
			case "count":
				dest[i] = &count
			case "date":
				date = new(time.Time)
				dest[i] = date
			case "user_agent":
				userAgent = new(string)
				dest[i] = userAgent
			default:
				return nil, fmt.Errorf("repository: unexpected column in result: %s", col)
			}
		}

		if err := rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("repository: %v", err)
		}

		results = append(results, models.TransitAggregationResult{
			Date:      date,
			UserAgent: userAgent,
			Count:     count,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: %v", err)
	}

	return results, nil
}
