package main

import (
	"net/http"
)

func main() {
	// TODO settings.cfg, launching of background tasks, global context?
	http.ListenAndServe(":9995", GetRouter())
}
