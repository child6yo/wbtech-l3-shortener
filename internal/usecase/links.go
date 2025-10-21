package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/child6yo/wbtech-l3-shortener/internal/models"
	"github.com/child6yo/wbtech-l3-shortener/internal/repository"
)

const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
const retries = 6
const startLinkLenght = 3

type linksRepository interface {
	CreateLink(ctx context.Context, link models.Link) error
	GetFullLink(ctx context.Context, short models.ShortLink) (models.FullLink, error)
}

// LinksShortener сокращает ссылки и сохраняет их в репозиторий.
type LinksShortener struct {
	repo linksRepository
}

// NewLinksShortener создает новый LinksShortener.
func NewLinksShortener(repo linksRepository) *LinksShortener {
	return &LinksShortener{repo: repo}
}

// AddLink добавляет новую ссылку. Если кастомной ссылки не предложено - генерирует.
func (ls *LinksShortener) AddLink(ctx context.Context, link models.Link) (models.ShortLink, error) {
	if link.Short == models.ShortLink("") {
		return ls.generateLink(ctx, link.Full)
	} else {
		return link.Short, ls.addLink(ctx, link)
	}
}

func (ls *LinksShortener) generateLink(ctx context.Context, link models.FullLink) (models.ShortLink, error) {
	linkLenght := startLinkLenght

	for range retries {
		shortenLink := generate(linkLenght)
		err := ls.repo.CreateLink(
			ctx, models.Link{Short: models.ShortLink(shortenLink), Full: link})
		if errors.Is(err, repository.ErrAlreadyExist) {
			linkLenght++
			continue
		} else if err != nil {
			return models.ShortLink(""), fmt.Errorf("usecase: %v", err)
		}

		return models.ShortLink(shortenLink), nil
	}

	return models.ShortLink(""), fmt.Errorf("usecase: %w", ErrCollision)
}

func (ls *LinksShortener) addLink(ctx context.Context, link models.Link) error {
	err := ls.repo.CreateLink(ctx, link)
	if errors.Is(err, repository.ErrAlreadyExist) {
		return fmt.Errorf("usecase: %w: %v", ErrAlreadyExist, err)
	} else if err != nil {
		return fmt.Errorf("usecase: %v", err)
	}

	return nil
}

func generate(length int) string {
	var state uint64 = 88172645463325252

	next := func() uint64 {
		state ^= state << 13
		state ^= state >> 7
		state ^= state << 17
		return state
	}

	const base = uint64(62)
	buf := make([]byte, length)
	for i := range buf {
		buf[i] = chars[next()%base]
	}
	return string(buf)
}
