package tg

import (
	"fmt"
	"strconv"

	tele "gopkg.in/telebot.v3"
)

type Client struct {
	bot    *tele.Bot
	chatID tele.ChatID
	chat   *tele.Chat
}

func NewClient(token string, chatID tele.ChatID) (*Client, error) {
	bot, err := tele.NewBot(tele.Settings{
		Token: token,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		bot:    bot,
		chatID: chatID,
		chat: &tele.Chat{
			ID: int64(chatID),
		},
	}, nil
}

func (c *Client) SendSpoilerLink(threadID int, header, link string) (int, error) {
	payload := fmt.Sprintf("%s\n%s", header, link)

	message, err := c.bot.Send(c.chatID, payload, &tele.SendOptions{
		ThreadID:              threadID,
		DisableWebPagePreview: true,
		Entities: []tele.MessageEntity{
			{
				Type:   tele.EntitySpoiler,
				Offset: len(header) + 1,
				Length: len(link),
			},
		},
	})
	if err != nil {
		return 0, err
	}

	return message.ID, nil
}

func (c *Client) SendSticker(threadID int, stickerID string) (int, error) {
	sticker := tele.Sticker{
		File: tele.File{
			FileID: stickerID,
		},
	}
	message, err := c.bot.Send(c.chatID, &sticker, &tele.SendOptions{
		ThreadID: threadID,
	})
	if err != nil {
		return 0, fmt.Errorf("send: %w", err)
	}

	return message.ID, nil
}

func (c *Client) Pin(id int) error {
	msg := tele.StoredMessage{
		MessageID: strconv.Itoa(id),
		ChatID:    c.chat.ID,
	}

	return c.bot.Pin(msg, tele.Silent)
}

func (c *Client) Unpin(id int) error {
	return c.bot.Unpin(c.chat, id)
}
