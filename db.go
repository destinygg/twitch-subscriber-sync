package main

import (
	"database/sql"

	"github.com/destinygg/website2/internal/config"
	"github.com/destinygg/website2/internal/debug"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/net/context"
)

func InitDB(ctx context.Context) context.Context {
	cfg, _ := ctx.Value("appconfig").(*config.AppConfig)

	var err error
	db, err := sql.Open("mysql", cfg.Database.DSN)
	if err != nil {
		d.F("Could not open database: %#v", err)
	}

	err = db.Ping()
	if err != nil {
		d.F("Could not connect to database: %#v", err)
	}

	db.SetMaxIdleConns(cfg.Database.MaxIdleConnections)
	db.SetMaxOpenConns(cfg.Database.MaxConnections)

	return context.WithValue(ctx, "db", db)
}
