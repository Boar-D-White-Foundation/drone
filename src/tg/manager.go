package tg

import (
	"fmt"

	tele "gopkg.in/telebot.v3"
)

const (
	DefaultDailyHeader = "LeetCode Daily Question"
)

type Manager struct {
	BoarDWhiteChatID           tele.ChatID
	BoarDWhiteLeetCodeThreadID int
}

func (m *Manager) SendLCDailyToBoarDWhite(bot *tele.Bot, header, dailyLink string) error {
	payload := fmt.Sprintf("%v\n%v", header, dailyLink)
	opts := tele.SendOptions{
		ThreadID:              m.BoarDWhiteLeetCodeThreadID,
		DisableWebPagePreview: true,
		Entities: []tele.MessageEntity{
			{
				Type:   tele.EntitySpoiler,
				Offset: len(header) + 1,
				Length: len(dailyLink),
			},
		},
	}
	_, err := bot.Send(m.BoarDWhiteChatID, payload, &opts)
	if err != nil {
		return err
	}
	fmt.Println("Published lc daily:", dailyLink)
	return nil
}
