package main

import (
	"github.com/destinygg/website2/internal/config"
	"github.com/destinygg/website2/internal/debug"
	"github.com/destinygg/website2/internal/redis"
	"github.com/destinygg/website2/internal/smtp"
	"golang.org/x/net/context"
)

func main() {
	ctx := context.Background()
	ctx = config.Init(ctx)
	ctx = smtp.Init(ctx)
	ctx = d.Init(ctx)
	ctx = rds.Init(ctx)
	ctx = InitDB(ctx)
	InitWebsite(ctx)
}
