package database

import (
	"agropal/config"
	"fmt"
	"strconv"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewPostgres(c *config.Config) *sqlx.DB {
	// Convert port string to integer
	port, err := strconv.Atoi(c.Database.Port)
	if err != nil {
		panic(fmt.Sprintf("invalid port: %v", err))
	}

	// Build PostgreSQL connection string
	connection := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		c.Database.Host,
		c.Database.User,
		c.Database.Password,
		c.Database.Dbname,
		port,
	)

	// Establish database connection
	db, err := sqlx.Connect("postgres", connection)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	return db
}
