package model

type UrlMapping struct {
	ID       int64  `json:"id" gorm:"primary_key"`
	ShortUrl string `json:"short_url"`
	LongUrl  string `json:"long_url"`
}
