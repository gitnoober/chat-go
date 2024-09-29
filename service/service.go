package service

import "database/sql"


type Service struct {
	mysqlDB     *sql.DB
}

func NewService(
	mysqlDB *sql.DB,
) *Service {

	svc := &Service{
		mysqlDB:          mysqlDB,
	}
	return svc
}
