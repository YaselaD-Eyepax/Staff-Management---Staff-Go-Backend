package db

import (
	"events-service/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(cfg *config.Config) *gorm.DB {
    dsn := "host=" + cfg.DBHost +
        " user=" + cfg.DBUser +
        " password=" + cfg.DBPass +
        " dbname=" + cfg.DBName +
        " port=" + cfg.DBPort +
        " sslmode=disable"

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(err)
    }

    return db
}
