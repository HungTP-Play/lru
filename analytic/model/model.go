package model

import "time"

type AnalyticRecord struct {
	ID            int       `gorm:"primaryKey,autoIncrement" json:"id"`
	ShortUrl      string    `json:"short_url"`
	OriginalUrl   string    `json:"original_url"`
	RedirectCount int       `json:"redirect_count"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	LatestAccess  time.Time `gorm:"autoUpdateTime" json:"latest_access"`
}
