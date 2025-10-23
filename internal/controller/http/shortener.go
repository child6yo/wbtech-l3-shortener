package httpctrl

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/child6yo/wbtech-l3-shortener/internal/models"
	"github.com/child6yo/wbtech-l3-shortener/internal/usecase"
	"github.com/wb-go/wbf/ginext"
)

type analitycs interface {
	AsyncNewTransit(transit models.Transit)
	GetAggregatedTransits(
		ctx context.Context, query models.TransitAggregationQuery) ([]models.TransitAggregationResult, error)
}

type shortener interface {
	AddLink(ctx context.Context, link models.Link) (models.ShortLink, error)
	GetFullLink(ctx context.Context, shortLink models.ShortLink) (models.FullLink, error)
}

// ShortenerController http контроллер сокращателя ссылок.
type ShortenerController struct {
	shr shortener
	ans analitycs
}

// NewShortenerController создает новый ShortenerController.
func NewShortenerController(shr shortener, ans analitycs) *ShortenerController {
	return &ShortenerController{shr: shr, ans: ans}
}

type createShortenLinkRequest struct {
	CustomShort string `json:"custom,omitempty" binding:"omitempty,min=1,max=255"`
	Full        string `json:"full" binding:"required,url"`
}

// Shorten обрабатывает POST /shorten — создание новой сокращённой ссылки.
func (sc *ShortenerController) Shorten(c *ginext.Context) {
	var req createShortenLinkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ginext.H{"error": "invalid request: " + err.Error()})
		_ = c.Error(fmt.Errorf("validation error: %v", err))
		return
	}

	c.Set("request", req)

	link := models.Link{
		Short: models.ShortLink(req.CustomShort),
		Full:  models.FullLink(req.Full),
	}
	shorten, err := sc.shr.AddLink(c.Request.Context(), link)
	if errors.Is(err, usecase.ErrAlreadyExist) {
		c.JSON(409, ginext.H{"error": "this short link already exists"})
		_ = c.Error(fmt.Errorf("already exists: %v", err))
		return
	} else if err != nil {
		c.JSON(500, ginext.H{"error": "failed to create link"})
		_ = c.Error(fmt.Errorf("create link: %v", err))
		return
	}

	c.JSON(200, ginext.H{"shorten_link": shorten})
}

// Redirect обрабатывает GET /s/{short_url} — переход по короткой ссылке.
func (sc *ShortenerController) Redirect(c *ginext.Context) {
	shortURL := c.Param("short_url")
	c.Set("request", shortURL)

	fullURL, err := sc.shr.GetFullLink(c.Request.Context(), models.ShortLink(shortURL))
	if err != nil {
		c.JSON(500, ginext.H{"error": "failed to redirect"})
		_ = c.Error(fmt.Errorf("get full link: %v", err))
		return
	}

	transit := models.Transit{
		Link:      models.ShortLink(shortURL),
		Timestamp: time.Now().UTC(),
		UserAgent: c.GetHeader("User-Agent"),
	}

	sc.ans.AsyncNewTransit(transit)

	c.Redirect(http.StatusMovedPermanently, string(fullURL))
}

// GetAnalytics обрабатывает GET /analytics/{short_url} — получение аналитики (число переходов, User-Agent, время переходов).
func (sc *ShortenerController) GetAnalytics(c *ginext.Context) {
	shortURL := c.Param("short_url")

	transitAggr := models.TransitAggregationQuery{
		Link: models.ShortLink(shortURL),

		GroupByDay:       queryExists(c, "group_by_day"),
		GroupByMonth:     queryExists(c, "group_by_month"),
		GroupByUserAgent: queryExists(c, "group_by_user_agent"),
	}

	c.Set("request", transitAggr)

	res, err := sc.ans.GetAggregatedTransits(c.Request.Context(), transitAggr)
	if err != nil {
		c.JSON(500, ginext.H{"error": "failed to get aggregated transits"})
		_ = c.Error(fmt.Errorf("get aggregated transits: %v", err))
		return
	}

	c.JSON(200, res)
}

func queryExists(c *ginext.Context, queryKey string) bool {
	_, ok := c.GetQuery(queryKey)
	return ok
}
