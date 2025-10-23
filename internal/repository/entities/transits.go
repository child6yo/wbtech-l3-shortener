package entities

import "time"

// Transit определяет сущность хранения переходов по ссылкам.
type Transit struct {
	Link      string    `db:"link"`
	Timestamp time.Time `db:"timestamp"`
	UserAgent string    `db:"user_agent"`
}
