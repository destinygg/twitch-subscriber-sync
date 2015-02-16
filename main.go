package main

import (
	"net/http"

	"github.com/destinygg/website2/internal/debug"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// TODO settings.cfg, launching of background tasks, global context?
	d.Init(d.EnableDebug)
	http.ListenAndServe(":9995", GetRouter())
}
