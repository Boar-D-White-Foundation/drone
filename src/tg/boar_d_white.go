package tg

import (
	"fmt"

	tele "gopkg.in/telebot.v3"
)

const (
	DefaultDailyHeader = "LeetCode Daily Question"

	boarDWhiteChatID           = tele.ChatID(-1001640461540)
	boarDWhiteLeetCodeThreadID = 10095
)

func SendLCDailyToBoarDWhite(bot *tele.Bot, header, dailyLink string) error {
	payload := fmt.Sprintf("%v\n%v", header, dailyLink)
	opts := tele.SendOptions{
		ThreadID: boarDWhiteLeetCodeThreadID,
		Entities: []tele.MessageEntity{
			{
				Type:   tele.EntitySpoiler,
				Offset: len(header) + 1,
				Length: len(dailyLink),
			},
		},
	}
	_, err := bot.Send(boarDWhiteChatID, payload, &opts)
	if err != nil {
		return err
	}
	fmt.Println("Published lc daily:", dailyLink)
	return nil
}
