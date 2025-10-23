package models

import "time"

// Transit определяет модель переходов по ссылкам.
type Transit struct {
	Link      ShortLink
	Timestamp time.Time
	UserAgent string
}

type TransitAggregationQuery struct {
	Link ShortLink

	GroupByDay       bool
	GroupByMonth     bool
	GroupByUserAgent bool
}

type TransitAggregationResult struct {
	Date      *time.Time `json:"date,omitempty"`
	UserAgent *string    `json:"user_agent,omitempty"`
	Count     uint64     `json:"count"`
}
