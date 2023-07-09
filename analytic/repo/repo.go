package repo

import (
	"github.com/HungTP-Play/lru/analytic/model"
	"github.com/HungTP-Play/lru/shared"
)

type AnalyticRepo struct {
	ConnectionString string
	DB               shared.PostgresDB
}

func NewAnalyticRepo(connectionString string) *AnalyticRepo {
	db := shared.NewPostgresDB(connectionString)
	db.Init()
	return &AnalyticRepo{
		ConnectionString: connectionString,
		DB:               *db,
	}
}

func (repo *AnalyticRepo) Close() error {
	return repo.DB.Close()
}

func (repo *AnalyticRepo) Add(record *model.AnalyticRecord) error {
	return repo.DB.Create(record)
}

func (repo *AnalyticRepo) IncAccessCount(shortUrl string) error {
	var record model.AnalyticRecord
	err := repo.DB.Find(&record, "short_url = ?", shortUrl)
	if err != nil {
		return err
	}

	record.RedirectCount += 1
	return repo.DB.DB.Save(&record).Error
}
