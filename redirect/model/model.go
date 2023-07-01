package model

type RedirectUrl struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Url      string `json:"url"`
	ShortUrl string `gorm:"unique,index" json:"shortUrl"`
}
