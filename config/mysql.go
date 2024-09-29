package config

import (
	"log"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
)


type DBConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Net string `json:"net"`
	Addr string	`json:"addr"`
	DBName string `json:"dbname"`
}

func loadDBConfig() *DBConfig {
	return &DBConfig{
		Username: "root",
		Password: "123456",
		Net: "tcp",
		Addr: "127.0.0.1:3306",
		DBName: "test",
	}
}

func ConnectMysql(
	cfg *Config,
	db *sql.DB,
) (*sql.DB, error) {
	mysqlConfig := mysql.Config{
		User: cfg.DBConfig.Username,
		Passwd: cfg.DBConfig.Password,
		Net: cfg.DBConfig.Net,
		Addr: cfg.DBConfig.Addr,
		DBName: cfg.DBConfig.DBName,
	}

	// Get a database handle
	var err error
	db, err = sql.Open("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		log.Fatalf("Error encountered while loading database: %v", err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatalf("Error encountered while pinging database: %v", pingErr)
	}
	fmt.Printf("Connected to DB!")
	return db, nil
}