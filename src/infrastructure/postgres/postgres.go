package postgres

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var postgres_db = os.Getenv("POSTGRES_DB")
var postgres_user = os.Getenv("POSTGRES_USER")
var postgres_password = os.Getenv("POSTGRES_PASSWORD")
var postgres_host = os.Getenv("POSTGRES_HOST")
var postgres_port = os.Getenv("POSTGRES_PORT")

func Connect() (*gorm.DB, error) {
	var dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		postgres_host, postgres_user, postgres_password, postgres_db, postgres_port)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func Disconnect(db *gorm.DB) (err error) {
	conn, err := db.DB()
	if err != nil {
		return err
	}
	err = conn.Close()
	return err
}
