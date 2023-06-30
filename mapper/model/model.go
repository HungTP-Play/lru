package model

type UrlMapping struct {
	ID       int64  `json:"id" gorm:"primary_key"`
	ShortUrl string `json:"short_url" gorm:"unique,index"`
	LongUrl  string `json:"long_url"  gorm:"unique,index"`
}
