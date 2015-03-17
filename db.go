/***
  This file is part of destinygg/website2.

  Copyright (c) 2015 Peter Sztan <sztanpet@gmail.com>

  destinygg/website2 is free software; you can redistribute it and/or modify it
  under the terms of the GNU Lesser General Public License as published by
  the Free Software Foundation; either version 3 of the License, or
  (at your option) any later version.

  destinygg/website2 is distributed in the hope that it will be useful, but
  WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
  Lesser General Public License for more details.

  You should have received a copy of the GNU Lesser General Public License
  along with destinygg/website2; If not, see <http://www.gnu.org/licenses/>.
***/

package main

import (
	"database/sql"

	"github.com/destinygg/website2/internal/config"
	"github.com/destinygg/website2/internal/debug"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/net/context"
)

func InitDB(ctx context.Context) context.Context {
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
