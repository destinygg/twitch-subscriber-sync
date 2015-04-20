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
	"time"

	"github.com/destinygg/website2/internal/config"
	"github.com/destinygg/website2/internal/debug"
	"github.com/destinygg/website2/internal/redis"
	"github.com/destinygg/website2/internal/smtp"
	"golang.org/x/net/context"
)

func main() {
	time.Local = time.UTC
	ctx := context.Background()
	ctx = config.Init(ctx)
	ctx = smtp.Init(ctx)
	ctx = d.Init(ctx)
	ctx = rds.Init(ctx)
	ctx = InitDB(ctx)
	InitWebsite(ctx)
}
