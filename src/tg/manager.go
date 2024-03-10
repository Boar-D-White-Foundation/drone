package tg

import (
	"fmt"
	"log/slog"

	tele "gopkg.in/telebot.v3"
)

const (
	DefaultDailyHeader = "LeetCode Daily Question"
)

type Manager struct {
	BoarDWhiteChatID           tele.ChatID
	BoarDWhiteLeetCodeThreadID int
	LCDailyStickerID           string
}

func (m *Manager) SendLCDailyToBoarDWhite(bot *tele.Bot, header, dailyLink string) error {
	payload := fmt.Sprintf("%v\n%v", header, dailyLink)
	_, err := bot.Send(m.BoarDWhiteChatID, payload, &tele.SendOptions{
		ThreadID:              m.BoarDWhiteLeetCodeThreadID,
		DisableWebPagePreview: true,
		Entities: []tele.MessageEntity{
			{
				Type:   tele.EntitySpoiler,
				Offset: len(header) + 1,
				Length: len(dailyLink),
			},
		},
	})
	if err != nil {
		return err
	}
	slog.Info("published lc daily", slog.String("link", dailyLink))

	sticker := tele.Sticker{File: tele.File{FileID: m.LCDailyStickerID}}
	_, err = bot.Send(m.BoarDWhiteChatID, &sticker, &tele.SendOptions{
		ThreadID: m.BoarDWhiteLeetCodeThreadID,
	})
	if err != nil {
		return err
	}
	slog.Info("published daily sticker", slog.String("id", m.LCDailyStickerID))
	return nil
}
