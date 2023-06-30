package repo

import (
	"os"

	"github.com/HungTP-Play/lru/mapper/model"
	"github.com/HungTP-Play/lru/mapper/util"
	"github.com/HungTP-Play/lru/shared"
)

type UrlMappingRepo struct {
	ConnectionString string
	DB               shared.PostgresDB
}

func NewUrlMappingRepo(connectionString string) *UrlMappingRepo {
	db := shared.NewPostgresDB(connectionString)
	db.Init()
	return &UrlMappingRepo{
		ConnectionString: connectionString,
		DB:               *db,
	}
}

func (repo *UrlMappingRepo) Close() error {
	return repo.DB.Close()
}

func (repo *UrlMappingRepo) Map(urlMappingRequest shared.MapUrlRequest) (string, error) {
	var urlMapping model.UrlMapping

	totalUrls, err := repo.DB.CountTable(&urlMapping)
	if err != nil {
		return "", err
	}

	stringEncode := util.Base62Encode(totalUrls + 1)

	baseHost := os.Getenv("BASE_HOST")
	if baseHost == "" {
		baseHost = "http://localhost/"
	}

	urlMapping = model.UrlMapping{
		ShortUrl: baseHost + stringEncode,
		LongUrl:  urlMappingRequest.Url,
	}

	err = repo.DB.Create(&urlMapping)
	if err != nil {
		return baseHost + stringEncode, err
	}
	return baseHost + stringEncode, err
}
