package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"

	"github.com/destinygg/website2/internal/config"
	"github.com/destinygg/website2/internal/debug"
	"github.com/destinygg/website2/internal/errorpages"
	"github.com/destinygg/website2/internal/user"
	"github.com/lionelbarrow/braintree-go"
	"golang.org/x/net/context"
)

func getBTFromContext(ctx context.Context) *braintree.Braintree {
	cfg := config.GetFromContext(ctx)
	return braintree.New(
		braintree.Environment(cfg.Braintree.Environment),
		cfg.Braintree.MerchantID,
		cfg.Braintree.PublicKey,
		cfg.Braintree.PrivateKey,
	)
}

// BraintreeVerifyHandler assumes its called for GET requests only and simply
// replies with the correct response expected the webhook for verification
// Takes the sha1 hash of the private key and uses it as the key in the
// hmac-sha1 of the bt_challenge query param, finally it outputs
// the publickey | hmac
func BraintreeVerifyHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	challenge := r.URL.Query().Get("bt_challenge")
	if challenge == "" {
		erp.BadRequest(w, r)
		return
	}

	cfg := config.GetFromContext(ctx)

	s := sha1.New()
	if _, err := io.WriteString(s, cfg.Braintree.PrivateKey); err != nil {
		panic(err.Error())
	}

	mac := hmac.New(sha1.New, s.Sum(nil))
	if _, err := mac.Write([]byte(challenge)); err != nil {
		panic(err.Error())
	}

	fmt.Fprintf(w, "%s|%x", cfg.Braintree.PublicKey, mac.Sum(nil))
}

func BraintreeHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	signature := r.PostFormValue("bt_signature")
	payload := r.PostFormValue("bt_payload")
	bt := getBTFromContext(ctx)

	wh := bt.WebhookNotification()
	not, err := wh.Parse(signature, payload)
	d.D(r, not, err)
	panic("not a 200, sorry, retry")
}

func GenerateClientToken() {
	// TODO needed for paypal, also paypal can only work on https
}

func CreateSubscription(u *user.User, typ string) {
	// TODO does the user have a guid? if not create a braintree customer (maybe with CC info?)
	// and update the user with the guid braintree generates for the user
	// then either create a transaction or a subscription that is recurring
	// signal in the transaction that its ?recurring? and to auto submit_for_settlement
	// and store_in_vault_on_success and the customer id
	// if paymentmethodnonce is present, its a paypal payment, identified by email, do we get that back?
}
