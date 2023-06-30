package repo

import "github.com/HungTP-Play/lru/shared"

type UrlMappingRepo struct {
	db shared.PostgresDB
}

func NewUrlMappingRepo(connectionString string) *UrlMappingRepo {
	db := shared.NewPostgresDB(connectionString)
	return &UrlMappingRepo{
		db: *db,
	}
}
