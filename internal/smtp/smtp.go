/***
  This file is part of destinygg/smtp.

  Copyright (c) 2015 Peter Sztan <sztanpet@gmail.com>

  destinygg/smtp is free software; you can redistribute it and/or modify it
  under the terms of the GNU Lesser General Public License as published by
  the Free Software Foundation; either version 3 of the License, or
  (at your option) any later version.

  destinygg/smtp is distributed in the hope that it will be useful, but
  WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
  Lesser General Public License for more details.

  You should have received a copy of the GNU Lesser General Public License
  along with destinygg/smtp; If not, see <http://www.gnu.org/licenses/>.
***/

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
