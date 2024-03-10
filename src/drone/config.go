package main

import (
	"errors"
	"os"
	"strconv"

	tele "gopkg.in/telebot.v3"
)

const (
	defaultStickerID = "CAACAgIAAxkBAAELpiFl7G4Rn8WQBK3AaDiAMn6ixTUR7gACzzkAAr-TAAFK91qMnVpp9TQ0BA"
	LCDPinBadgerKey  = "lcd-pinned-message"
)

type Config struct {
	TgKey                      string
	LCDailyCron                string
	LCDailyStickerID           string
	BoarDWhiteChatID           tele.ChatID
	BoarDWhiteLeetCodeThreadID int
	BadgerPath                 string
}

func LoadConfig() (Config, error) {
	boarDWhiteChatID, err := getEnvInt64Default("DRONE_BOAR_D_WHITE_CHAT_ID", -1001640461540)
	if err != nil {
		return Config{}, errors.New("chat id is incorrect")
	}

	boarDWhiteLeetCodeThreadID, err := getEnvIntDefault("DRONE_BOAR_D_WHITE_LEET_CODE_THREAD_ID", 10095)
	if err != nil {
		return Config{}, errors.New("thread id is incorrect")
	}

	return Config{
		TgKey:                      os.Getenv("DRONE_TG_BOT_API_KEY"),
		LCDailyCron:                getEnvDefault("DRONE_LC_DAILY_CRON", "0 * * * *"), // every hour
		LCDailyStickerID:           getEnvDefault("DRONE_LC_DAILY_STICKER_ID", defaultStickerID),
		BoarDWhiteChatID:           tele.ChatID(boarDWhiteChatID),
		BoarDWhiteLeetCodeThreadID: boarDWhiteLeetCodeThreadID,
		BadgerPath:                 getEnvDefault("DRONE_BADGER_PATH", "badger"), // default to relative path inside the working dir
	}, nil
}

func getEnvDefault(key, value string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return value
	}
	return v
}

func getEnvIntDefault(key string, value int) (int, error) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return value, nil
	}
	return strconv.Atoi(v)
}

func getEnvInt64Default(key string, value int64) (int64, error) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return value, nil
	}
	return strconv.ParseInt(v, 10, 64)
}
