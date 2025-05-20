package dbrepo

import (
	"database/sql"

	"github.com/racw/bookings/internal/config"
	"github.com/racw/bookings/internal/repository"
)

type postgresDBRepo struct {
	App *config.AppConfig
	DB *sql.DB
}

type mysqlDBRepo struct {
	App *config.AppConfig
	DB *sql.DB
}

type testDBRepo struct {
	App *config.AppConfig
	DB *sql.DB
}

func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &postgresDBRepo{
		App: a,
		DB: conn,
	}
}

func NewTestingRepo(a *config.AppConfig) repository.DatabaseRepo {
	return &testDBRepo{
		App: a,
	}
}

// func NewMysqlRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
// 	return &mysqlDBRepo{
// 		App: a,
// 		DB: conn,
// 	}
// }