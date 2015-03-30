package db

import (
	"database/sql"

	"github.com/destinygg/website2/internal/config"
	"github.com/destinygg/website2/internal/debug"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/net/context"
)

func Init(ctx context.Context) context.Context {
	cfg := config.GetFromContext(ctx)
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

func GetFromContext(ctx context.Context) *sql.DB {
	db, ok := ctx.Value("db").(*sql.DB)
	if !ok {
		panic("SQL database not found in the context")
	}

	return db
}
