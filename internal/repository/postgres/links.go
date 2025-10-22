package postgres

import (
	"context"
	"fmt"

	"github.com/child6yo/wbtech-l3-shortener/internal/models"
	"github.com/child6yo/wbtech-l3-shortener/internal/repository"
	"github.com/lib/pq"
	"github.com/wb-go/wbf/dbpg"
)

// LinksRepository отвечает за область хранения сокращенных ссылок в Postgres.
type LinksRepository struct {
	db *dbpg.DB
}

// NewLinksRepository создает новый экземпляр LinksRepository.
func NewLinksRepository(db *dbpg.DB) *LinksRepository {
	return &LinksRepository{db: db}
}

// CreateLink создает новую ссылку в базе данных.
func (lr *LinksRepository) CreateLink(ctx context.Context, link models.Link) error {
	query := fmt.Sprintf("INSERT INTO %s (short, destination) VALUES ($1, $2)", tableLinks)

	_, err := lr.db.ExecContext(ctx, query, link.Short, link.Full)
	if isUniqueViolation(err) {
		return fmt.Errorf("repository: %w: %v", repository.ErrAlreadyExist, err)
	} else if err != nil {
		return fmt.Errorf("repository: %v", err)
	}

	return nil
}

// GetFullLink возвращает полную ссылку по сокращенной.
func (lr *LinksRepository) GetFullLink(ctx context.Context, short models.ShortLink) (models.FullLink, error) {
	query := fmt.Sprintf("SELECT destination FROM %s WHERE short=$1", tableLinks)

	row := lr.db.QueryRowContext(ctx, query, short)

	var res string
	err := row.Scan(&res)

	return models.FullLink(res), err
}

func isUniqueViolation(err error) bool {
	pqErr, ok := err.(*pq.Error)
	return ok && pqErr.Code == "23505"
}
