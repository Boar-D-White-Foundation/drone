package main

import (
	"errors"
	"os"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"
)

const (
	defaultLCStickerID = "CAACAgIAAxkBAAELpiFl7G4Rn8WQBK3AaDiAMn6ixTUR7gACzzkAAr-TAAFK91qMnVpp9TQ0BA"
	defaultNCStickerID = "CAACAgIAAxkBAAELqk5l72yJkbx4_vskG3n6zWoWaAnA3QACazYAArJX2UsY5inoNwaFoTQE"
)

type Config struct {
	TgKey                      string
	LCDailyCron                string
	LCDailyStickerID           string
	NCDailyCron                string
	NCDailyStickerID           string
	NCDailyStartDate           time.Time
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

	ncDailyStartDate, err := getEnvTimeDefault(
		"DRONE_NC_DAILY_START_DATE",
		time.Date(2024, 3, 14, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		return Config{}, errors.New("nc daily start date is incorrect")
	}

	return Config{
		TgKey: os.Getenv("DRONE_TG_BOT_API_KEY"),
		// every day at 01:00 UTC
		LCDailyCron:      getEnvDefault("DRONE_LC_DAILY_CRON", "0 1 * * *"),
		LCDailyStickerID: getEnvDefault("DRONE_LC_DAILY_STICKER_ID", defaultLCStickerID),
		// every day at 13:00 UTC
		NCDailyCron:                getEnvDefault("DRONE_NC_DAILY_CRON", "0 13 * * *"),
		NCDailyStickerID:           getEnvDefault("DRONE_NC_DAILY_STICKER_ID", defaultNCStickerID),
		NCDailyStartDate:           ncDailyStartDate,
		BoarDWhiteChatID:           tele.ChatID(boarDWhiteChatID),
		BoarDWhiteLeetCodeThreadID: boarDWhiteLeetCodeThreadID,
		// default to relative path inside the working dir
		BadgerPath: getEnvDefault("DRONE_BADGER_PATH", "data/badger"),
	}, nil
}

func getEnvDefault(key, value string) string {
	v := os.Getenv(key)
	if len(v) == 0 {
		return value
	}
	return v
}

func getEnvIntDefault(key string, value int) (int, error) {
	v := os.Getenv(key)
	if len(v) == 0 {
		return value, nil
	}
	return strconv.Atoi(v)
}

func getEnvInt64Default(key string, value int64) (int64, error) {
	v := os.Getenv(key)
	if len(v) == 0 {
		return value, nil
	}
	return strconv.ParseInt(v, 10, 64)
}

func getEnvTimeDefault(key string, value time.Time) (time.Time, error) {
	v := os.Getenv(key)
	if len(v) == 0 {
		return value, nil
	}
	return time.Parse("2006-01-02", v)
}
