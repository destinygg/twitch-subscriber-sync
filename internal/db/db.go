package db

import (
	"github.com/destinygg/website2/internal/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
)

func Init(ctx context.Context) context.Context {
	cfg := config.GetFromContext(ctx)
	db := sqlx.MustConnect("mysql", cfg.Database.DSN)

	db.SetMaxIdleConns(cfg.Database.MaxIdleConnections)
	db.SetMaxOpenConns(cfg.Database.MaxConnections)

	return context.WithValue(ctx, "db", db)
}

func GetFromContext(ctx context.Context) *sqlx.DB {
	db, ok := ctx.Value("db").(*sqlx.DB)
	if !ok {
		panic("SQL database not found in the context")
	}

	return db
}
