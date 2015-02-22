package smtp

import (
	//"bytes"
	//"net/smtp"
	"crypto/tls"
	"strconv"
	"strings"
	"text/template"

	"github.com/destinygg/website2/internal/config"
	"github.com/destinygg/website2/internal/debug"
	"golang.org/x/net/context"
	"gopkg.in/gomail.v1"
)

var (
	mailer *gomail.Mailer
	conf   config.SMTP
)

func Init(ctx context.Context) context.Context {
	cfg := config.GetFromContext(ctx)
	conf = cfg.SMTP

	p := strings.Index(conf.Addr, ":")
	if p < 0 {
		return ctx
	}

	host := conf.Addr[0:p]
	port, err := strconv.Atoi(conf.Addr[p+1:])
	if err != nil {
		d.FBT(port, err)
	}

	mailer = gomail.NewMailer(
		host,
		conf.Username,
		conf.Password,
		port,
		gomail.SetTLSConfig(&tls.Config{}),
	)

	return ctx
}

func Send(ctx context.Context, backtrace string) {
	if mailer == nil { // means it was not set up, so just bail out
		return
	}

	// TODO fix subjects, templates
	msg := gomail.NewMessage()
	msg.SetHeader("From", conf.FromEmail)
	msg.SetHeader("To", conf.LogEmail...)
	msg.SetHeader("Subject", "[destinygg] error TODO")

	w := msg.GetBodyWriter("text/plain")
	t := template.Must(template.New("errormail").Parse("TODO {{.}}"))
	t.Execute(w, backtrace)
	err := mailer.Send(msg)
	if err != nil {
		d.D(err)
	}
}
