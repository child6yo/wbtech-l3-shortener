package httpctrl

import (
	"context"
	"errors"
	"fmt"

	"github.com/child6yo/wbtech-l3-shortener/internal/models"
	"github.com/child6yo/wbtech-l3-shortener/internal/usecase"
	"github.com/wb-go/wbf/ginext"
)

type shortener interface {
	AddLink(ctx context.Context, link models.Link) (models.ShortLink, error)
}

// ShortenerController http контроллер сокращателя ссылок.
type ShortenerController struct {
	shr shortener
}

// NewShortenerController создает новый ShortenerController.
func NewShortenerController(shr shortener) *ShortenerController {
	return &ShortenerController{shr: shr}
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
		_ = c.Error(fmt.Errorf("validation error: %w", err))
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
		_ = c.Error(fmt.Errorf("already exists: %w", err))
		return
	} else if err != nil {
		c.JSON(500, ginext.H{"error": "failed to create link"})
		_ = c.Error(fmt.Errorf("failed to create link: %w", err))
		return
	}

	c.JSON(200, ginext.H{"shorten_link": shorten})
}
