package main

import (
	"net/http"

	"github.com/destinygg/website2/internal/debug"
	"golang.org/x/net/context"
)

func DonationHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	d.D(r)
}
