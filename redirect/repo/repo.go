package repo

import (
	"github.com/HungTP-Play/lru/redirect/model"
	"github.com/HungTP-Play/lru/shared"
)

type RedirectUrlRepo struct {
	ConnectionString string
	DB               shared.PostgresDB
}

func NewRedirectUrlRepo(connectionString string) *RedirectUrlRepo {
	db := shared.NewPostgresDB(connectionString)
	db.Init()
	return &RedirectUrlRepo{
		ConnectionString: connectionString,
		DB:               *db,
	}
}

func (repo *RedirectUrlRepo) Close() error {
	return repo.DB.Close()
}

func (repo *RedirectUrlRepo) AddRedirect(message shared.RedirectMessage) error {
	redirectUrl := model.RedirectUrl{
		ShortUrl: message.Shorten,
		Url:      message.Url,
	}

	err := repo.DB.Create(&redirectUrl)
	if err != nil {
		return err
	}
	return nil
}

func (repo *RedirectUrlRepo) GetRedirect(shorten string) (string, error) {
	var redirectUrl model.RedirectUrl
	err := repo.DB.Find(&redirectUrl, "short_url = ?", shorten)
	if err != nil {
		return "", err
	}

	return redirectUrl.Url, nil
}
