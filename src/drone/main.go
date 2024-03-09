package main

import (
	_ "go.uber.org/automaxprocs"

	"context"
	"os"
	"os/signal"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg := Config{
		TgKey:       os.Getenv("DRONE_TG_BOT_API_KEY"),
		LCDailyCron: os.Getenv("DRONE_LC_DAILY_CRON"),
	}
	if err := StartDrone(ctx, cfg); err != nil {
		panic(err)
	}
}
