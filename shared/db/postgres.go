package db

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresDB struct {
	ConnectionString string
	DB               *gorm.DB
}

func GetConnectionString() string {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	return "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=disable"
}

func NewPostgresDB(connectionString string) *PostgresDB {
	if connectionString == "" {
		connectionString = GetConnectionString()
	}
	return &PostgresDB{
		ConnectionString: connectionString,
	}
}

func (db *PostgresDB) Init() error {

	gdb, err := gorm.Open(postgres.Open(db.ConnectionString), &gorm.Config{})
	if err != nil {
		return err
	}

	db.DB = gdb
	return nil
}

func (db *PostgresDB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	sqlDB.Close()
	return nil
}

func (db *PostgresDB) GetDB() *gorm.DB {
	return db.DB
}

func (db *PostgresDB) Migrate(model interface{}) error {
	err := db.DB.AutoMigrate(model)
	return err
}

func (db *PostgresDB) Create(model interface{}) error {
	result := db.DB.Create(model)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (db *PostgresDB) Find(model interface{}, query interface{}, args ...interface{}) error {
	result := db.DB.Where(query, args...).Find(model)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (db *PostgresDB) Count(model interface{}, query interface{}, args ...interface{}) (int64, error) {
	var count int64
	result := db.DB.Model(model).Where(query, args...).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}

func (db *PostgresDB) CountTable(model interface{}) (int64, error) {
	var count int64
	result := db.DB.Model(model).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}
