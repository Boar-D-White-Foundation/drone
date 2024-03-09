package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"

	_ "go.uber.org/automaxprocs"
	tele "gopkg.in/telebot.v3"
)

func must(err error) {
	if err == nil {
		return
	}

	_, _ = fmt.Fprint(os.Stderr, err.Error())
	os.Exit(1)
}

func parseEnvInt(key string) (int, error) {
	val := os.Getenv(key)
	if len(val) == 0 {
		return 0, nil
	}
	return strconv.Atoi(val)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	boarDWhiteChatID, err := parseEnvInt("DRONE_BOAR_D_WHITE_CHAT_ID")
	must(err)
	boarDWhiteLeetCodeThreadID, err := parseEnvInt("DRONE_BOAR_D_WHITE_LEET_CODE_THREAD_ID")
	must(err)
	cfg := Config{
		TgKey:                      os.Getenv("DRONE_TG_BOT_API_KEY"),
		LCDailyCron:                os.Getenv("DRONE_LC_DAILY_CRON"),
		BoarDWhiteChatID:           tele.ChatID(boarDWhiteChatID),
		BoarDWhiteLeetCodeThreadID: boarDWhiteLeetCodeThreadID,
	}
	if len(cfg.LCDailyCron) == 0 {
		cfg.LCDailyCron = "0 1 * * *" // every day at 01:00 UTC
	}
	if cfg.BoarDWhiteChatID == 0 {
		cfg.BoarDWhiteChatID = tele.ChatID(-1001640461540)
	}
	if cfg.BoarDWhiteLeetCodeThreadID == 0 {
		cfg.BoarDWhiteLeetCodeThreadID = 10095
	}

	err = StartDrone(ctx, cfg)
	must(err)
}
