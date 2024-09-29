package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

type DBConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Net      string `json:"net"`
	Addr     string `json:"addr"`
	DBName   string `json:"dbname"`
}

func loadDBConfig() *DBConfig {
	return &DBConfig{
		Username: os.Getenv("MYSQL_USER"),
		Password: os.Getenv("MYSQL_PASSWORD"),
		Net:      "tcp",
		Addr:     "db:3306",
		DBName:   os.Getenv("MYSQL_DATABASE"),
	}
}

func ConnectMysql(
	cfg *Config,
	db *sql.DB,
) (*sql.DB, error) {
	mysqlConfig := mysql.Config{
		User:   cfg.DBConfig.Username,
		Passwd: cfg.DBConfig.Password,
		Net:    cfg.DBConfig.Net,
		Addr:   cfg.DBConfig.Addr,
		DBName: cfg.DBConfig.DBName,
	}

	// Get a database handle
	var err error
	dsn := mysqlConfig.FormatDSN()

	for i := 0; i < 5; i++ {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Printf("Error encountered while loading database: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if err = db.Ping(); err == nil {
			log.Println("Successfully connected to the database")
			return db, nil
		}

		log.Printf("Failed to ping database: %v", err)
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("could not connect to database after retries: %v", err)
}
