package usecase

import (
	"context"
	"time"

	"github.com/child6yo/wbtech-l3-shortener/internal/logger"
	"github.com/child6yo/wbtech-l3-shortener/internal/models"
)

type transitsRepository interface {
	CreateTransit(ctx context.Context, transit models.Transit) error
	GetAggregatedTransits(
		ctx context.Context,
		query models.TransitAggregationQuery,
	) ([]models.TransitAggregationResult, error)
}

// AnalyticsManager выдает аналитику по использованию сокращенных ссылок.
type AnalyticsManager struct {
	trRepo transitsRepository
	logger logger.Logger
}

// NewAnalyticsManager создает новый AnalyticsManager.
func NewAnalyticsManager(trRepo transitsRepository) *AnalyticsManager {
	return &AnalyticsManager{trRepo: trRepo}
}

// AsyncNewTransit асинхронно регистрирует новый переход по ссылке.
func (am *AnalyticsManager) AsyncNewTransit(transit models.Transit) {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	go func() {
		defer cancel()

		if err := am.trRepo.CreateTransit(ctxWithTimeout, transit); err != nil {
			am.logger.WithFields("transit", transit).Error(err)
		}
	}()
}

// GetAggregatedTransits количество переходов, агрегированных по query.
// Если в query не прописаны агрегации - агрегирует по user_agent.
func (am *AnalyticsManager) GetAggregatedTransits(
	ctx context.Context, query models.TransitAggregationQuery) ([]models.TransitAggregationResult, error) {
	return am.trRepo.GetAggregatedTransits(ctx, query)
}
