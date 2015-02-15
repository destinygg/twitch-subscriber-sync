package main

import (
	"net/http"

	"github.com/destinygg/website2/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

func GetRouter() *mux.Router {
	r := mux.NewRouter()
	chain := alice.New(middleware.Recover)
	authedchain := chain.Append(middleware.Auth)

	r.Handle("/donate", authedchain.ThenFunc(DonationHandler))
	http.Handle("/", r)
	return r
}
