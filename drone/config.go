package main

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"
)

const (
	defaultDPStickerID = "CAACAgIAAxkBAAELsFhl8hGciGxpkKi4-jhou97SOqwkvwACpT0AAkV6sEpqc1XNPnvOIDQE"
)

var (
	defaultDailyStickersIDs = []string{
		"CAACAgIAAxkBAAELpiFl7G4Rn8WQBK3AaDiAMn6ixTUR7gACzzkAAr-TAAFK91qMnVpp9TQ0BA",
		"CAACAgIAAxkBAAELqk5l72yJkbx4_vskG3n6zWoWaAnA3QACazYAArJX2UsY5inoNwaFoTQE",
		"CAACAgIAAxkBAAELsFZl8hFWX-2VvyppubPYPnLFrVeLPgACZjEAAmSSoUrAsMs8kC1omDQE",
		"CAACAgIAAxkBAAELsFRl8hE8RstoYI8E3RfKk2N1LRX-qQAC4TUAAhJs2ElFZyKM69wJfDQE",
		"CAACAgIAAxkBAAELsFJl8hEuJlU_1RK4nbpq3_iuZYsAATAAAi01AALtZihJuuf3iNlNcbU0BA",
		"CAACAgIAAxkBAAELsFBl8hEnQ9Ie_ANBN_TfzbDjvMck2wACVToAAhqdoEiRsQbJLo-jijQE",
		"CAACAgIAAxkBAAELsE5l8hEfhr-0Dn7BfsI9sDWA3I0t0wACSzUAAtS3aUj_OJb99K6DkjQE",
		"CAACAgIAAxkBAAELsExl8hEZTP0DFsf6hUDiZnx-i9I25wACpTEAAlp3YUhJNRORPRx_5DQE",
		"CAACAgIAAxkBAAELsEpl8hERg5mpamVfTo_SCgZRatbi6wACCzcAAqVsKEjLAvd1EuqdLzQE",
		"CAACAgIAAxkBAAELsEhl8hENAzygO8iFauBYZD0XPYqD3gACkTQAArnoCUjvaNgUd-BoHDQE",
		"CAACAgIAAxkBAAELsEZl8hD-i6wYaeLUMtP8MWhZwZoy3gACAjMAAs0gAUjsT0apWa4cRTQE",
		"CAACAgIAAxkBAAELsERl8hD1xsuLUYRD9F4a8ekVAgg8VAACgDgAAsuH-UvwActGs5DfMzQE",
		"CAACAgIAAxkBAAELsEJl8hDUc-b0jyDfeH6Ct2McMp4mlAACOzcAAquEmUsMP7ObsCcumTQE",
	}
)

type Config struct {
	TgKey                      string
	LCDailyCron                string
	NCDailyCron                string
	NCDailyStartDate           time.Time
	DailyStickerIDs            []string
	DPStickerID                string
	BoarDWhiteChatID           tele.ChatID
	BoarDWhiteLeetCodeThreadID int
	BadgerPath                 string
	TgPollerTimeout            int
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

	dailyStickerIDs, err := getEnvStringSliceDefault(",", "DRONE_DAILY_STICKER_IDS", defaultDailyStickersIDs)
	if err != nil {
		return Config{}, errors.New("daily sticker id's is incorrect")
	}

	tgPollerTimeout, err := getEnvIntDefault("DRONE_TG_POLLER_TIMEOUT", 10)
	if err != nil {
		return Config{}, errors.New("tg poller timeout is incorrect")
	}

	return Config{
		TgKey: os.Getenv("DRONE_TG_BOT_API_KEY"),
		// every day at 01:00 UTC
		LCDailyCron: getEnvDefault("DRONE_LC_DAILY_CRON", "0 1 * * *"),
		// every day at 12:00 UTC
		NCDailyCron:                getEnvDefault("DRONE_NC_DAILY_CRON", "0 12 * * *"),
		NCDailyStartDate:           ncDailyStartDate,
		DailyStickerIDs:            dailyStickerIDs,
		DPStickerID:                getEnvDefault("DRONE_DP_STICKER_ID", defaultDPStickerID),
		BoarDWhiteChatID:           tele.ChatID(boarDWhiteChatID),
		BoarDWhiteLeetCodeThreadID: boarDWhiteLeetCodeThreadID,
		// default to relative path inside the working dir
		BadgerPath:      getEnvDefault("DRONE_BADGER_PATH", "data/badger"),
		TgPollerTimeout: tgPollerTimeout,
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

func getEnvStringSliceDefault(sep, key string, value []string) ([]string, error) {
	v := os.Getenv(key)
	if len(v) == 0 {
		return value, nil
	}
	return strings.Split(v, sep), nil
}
