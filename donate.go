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
	"net/http"

	"github.com/destinygg/website2/internal/debug"
	"golang.org/x/net/context"
)

func DonationHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	d.D(r)
}
