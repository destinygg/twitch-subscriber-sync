/***
  This file is part of destinygg/errorpages.

  Copyright (c) 2015 Peter Sztan <sztanpet@gmail.com>

  destinygg/errorpages is free software; you can redistribute it and/or modify it
  under the terms of the GNU Lesser General Public License as published by
  the Free Software Foundation; either version 3 of the License, or
  (at your option) any later version.

  destinygg/errorpages is distributed in the hope that it will be useful, but
  WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
  Lesser General Public License for more details.

  You should have received a copy of the GNU Lesser General Public License
  along with destinygg/errorpages; If not, see <http://www.gnu.org/licenses/>.
***/

package erp

import (
	"fmt"
	"net/http"

	"github.com/destinygg/website2/internal/debug"
)

func AuthRequired(w http.ResponseWriter, r *http.Request) {
	// TODO with a proper template and shit
	http.Error(w, "authentication required", http.StatusForbidden)
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	// TODO proper template again
	http.Error(w, "bad parameters", http.StatusBadRequest)
}

func Recover(reason interface{}, w http.ResponseWriter, r *http.Request) {
	// TODO get a stack trace and as much info as possible, save it under some
	// key and show the user only that key
	// also possibly email about the error
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)

	err := d.NewErrorTrace(5, reason)
	fmt.Fprint(w, err.Error())
}
