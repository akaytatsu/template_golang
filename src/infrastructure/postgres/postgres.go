package postgres

import (
	"app/entity"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var gormDB *gorm.DB

func Connect() *gorm.DB {

	if gormDB == nil {
		return conn()
	}

	return gormDB
}

func Migrations() {
	db := Connect()

	db.AutoMigrate(&entity.EntityUser{})
}

func conn() *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
	)

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	gormDB = conn

	return gormDB
}
