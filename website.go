package main

import (
	"net/http"

	"github.com/destinygg/website2/internal/config"
	"github.com/destinygg/website2/internal/debug"
	"github.com/destinygg/website2/internal/middleware"
	"github.com/guregu/kami"
	"golang.org/x/net/context"
)

// InitWebsite does not return
func InitWebsite(ctx context.Context) {
	// set up defaults
	kami.Context = ctx
	middleware.RegisterPanicHandler()

	// /donate requires authentication
	kami.Use("/donate", middleware.Auth)
	kami.Get("/donate", DonationHandler)
	kami.Post("/donate", DonationHandler)

	// braintree integration endpoints
	kami.Get("/braintree", BraintreeVerifyHandler)
	kami.Post("/braintree", BraintreeHandler)

	http.Handle("/", kami.Handler())

	cfg := config.GetFromContext(ctx)
	if err := http.ListenAndServe(cfg.Website.Addr, nil); err != nil {
		d.F("Failed to init website: %+v", err)
	}
}
