package tg

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"
)

type Client interface {
	SendSpoilerLink(threadID int, header, link string) (int, error)
	SendSticker(threadID int, stickerID string) (int, error)
	ReplyWithSticker(messageID int, stickerID string) (int, error)
	Pin(id int) error
	Unpin(id int) error
	SetReaction(messageID int, reaction Reaction, isBig bool) error
}

type Service struct {
	bot      *tele.Bot
	chatID   tele.ChatID
	chat     *tele.Chat
	handlers map[string][]handler
}

var _ Client = (*Service)(nil)

func NewService(token string, chatID tele.ChatID, longPollerTimeout time.Duration) (*Service, error) {
	poller := tele.LongPoller{
		Timeout: longPollerTimeout,
	}
	bot, err := tele.NewBot(tele.Settings{
		Token:       token,
		Poller:      &poller,
		Synchronous: true, // to ease of debug and avoid race conditions on data dependent updates
	})
	if err != nil {
		return nil, err
	}

	chat := tele.Chat{
		ID: int64(chatID),
	}
	return &Service{
		bot:      bot,
		chatID:   chatID,
		chat:     &chat,
		handlers: make(map[string][]handler),
	}, nil
}

type handler struct {
	name string
	f    tele.HandlerFunc
}

func (s *Service) RegisterHandler(endpoint string, name string, f tele.HandlerFunc) {
	s.handlers[endpoint] = append(s.handlers[endpoint], handler{
		name: name,
		f:    f,
	})
}

func wrapErrors(h handler) func(tele.Context) {
	return func(c tele.Context) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic in tg handler", slog.String("name", h.name), slog.Any("err", err))
			}
		}()
		err := h.f(c)
		if err != nil {
			slog.Error("err in tg handler", slog.String("name", h.name), slog.Any("err", err))
		}
	}
}

func (s *Service) Start() {
	for endpoint, handlers := range s.handlers {
		h := func(tc tele.Context) error {
			for _, h := range handlers {
				wrapErrors(h)(tc)
			}
			return nil
		}
		s.bot.Handle(endpoint, h)
	}

	go s.bot.Start()
}

func (s *Service) Stop() {
	s.bot.Stop()
}

func (s *Service) SendSpoilerLink(threadID int, header, link string) (int, error) {
	payload := fmt.Sprintf("%s\n%s", header, link)
	message, err := s.bot.Send(s.chatID, payload, &tele.SendOptions{
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

func (s *Service) SendSticker(threadID int, stickerID string) (int, error) {
	sticker := tele.Sticker{
		File: tele.File{
			FileID: stickerID,
		},
	}
	message, err := s.bot.Send(s.chatID, &sticker, &tele.SendOptions{
		ThreadID: threadID,
	})
	if err != nil {
		return 0, fmt.Errorf("send: %w", err)
	}

	return message.ID, nil
}

func (s *Service) ReplyWithSticker(messageID int, stickerID string) (int, error) {
	sticker := tele.Sticker{
		File: tele.File{
			FileID: stickerID,
		},
	}
	message, err := s.bot.Send(s.chatID, &sticker, &tele.SendOptions{
		ReplyTo: &tele.Message{
			ID: messageID,
		},
	})
	if err != nil {
		return 0, fmt.Errorf("send: %w", err)
	}

	return message.ID, nil
}

func (s *Service) Pin(id int) error {
	msg := tele.StoredMessage{
		MessageID: strconv.Itoa(id),
		ChatID:    s.chat.ID,
	}

	return s.bot.Pin(msg, tele.Silent)
}

func (s *Service) Unpin(id int) error {
	return s.bot.Unpin(s.chat, id)
}

type setMessageReactionReq struct {
	ChatID    tele.ChatID `json:"chat_id"`
	MessageID int         `json:"message_id"`
	Reactions []Reaction  `json:"reaction"`
	IsBig     bool        `json:"is_big,omitempty"`
}

func (s *Service) SetReaction(messageID int, reaction Reaction, isBig bool) error {
	reactionOptions := setMessageReactionReq{
		ChatID:    s.chatID,
		MessageID: messageID,
		// currently, as non-premium users, bots can set up to one reaction per message
		Reactions: []Reaction{reaction},
		IsBig:     isBig,
	}
	_, err := s.bot.Raw("setMessageReaction", reactionOptions)
	if err != nil {
		return fmt.Errorf("set reaction: %w", err)
	}

	return nil
}
