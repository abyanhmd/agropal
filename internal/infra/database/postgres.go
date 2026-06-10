package database

import (
	"agropal/config"
	"fmt"
	"strconv"

	"github.com/jmoiron/sqlx"
)

func NewPostgres(c *config.Config) *sqlx.DB {
	port, errConvert := strconv.Atoi(c.Database.Port)
	if errConvert != nil {
		panic(errConvert)
	}

	connection := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s sslmode=disable",
		c.Database.Host,
		port,
		c.Database.User,
		c.Database.Dbname,
	)

	db, errOpen := sqlx.Connect("postgres", connection)
	if errOpen != nil {
		panic(errOpen)
	}

	return db
}
