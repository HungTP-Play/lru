package model

type UrlMapping struct {
	ID       int64  `gorm:"primary_key" json:"id"`
	ShortUrl string `gorm:"index" json:"short_url" `
	LongUrl  string `gorm:"index" json:"long_url"`
}
