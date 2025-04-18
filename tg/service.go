package tg

import (
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/boar-d-white-foundation/drone/alert"
	"github.com/boar-d-white-foundation/drone/config"
	"github.com/boar-d-white-foundation/drone/iterx"
	"gopkg.in/telebot.v3"
	tele "gopkg.in/telebot.v3"
)

type Client interface {
	BotID() int64
	SendMonospace(threadID int, text string) (int, error)
	SendMarkdownV2(threadID int, text string) (int, error)
	SendText(threadID int, text string) (int, error)
	SendSpoilerLink(threadID int, header, link string) (int, error)
	SendSticker(threadID int, stickerID string) (int, error)
	ReplyWithSticker(messageID int, stickerID string) (int, error)
	ReplyWithSpoilerPhoto(messageID int, caption, name, mime string, reader io.ReadSeeker) (int, error)
	ReplyWithDocument(messageID int, name, mime string, reader io.ReadSeeker) (int, error)
	ReplyWithText(messageID int, text string) (int, error)
	EditMessageText(messageID int, text string) error
	Pin(id int) error
	Unpin(id int) error
	SetReaction(messageID int, reaction Reaction, isBig bool) error
	Delete(id int) error
}

type AdminClient interface {
	Client

	SendAlert(msg string) error
}

type HandlerRegistry interface {
	RegisterHandler(endpoint string, name string, f tele.HandlerFunc)
}

type Service struct {
	alerts   *alert.Manager
	bot      *tele.Bot
	chatID   tele.ChatID
	chat     *tele.Chat
	handlers map[string][]handler
}

var _ Client = (*Service)(nil)
var _ HandlerRegistry = (*Service)(nil)

func NewService(alerts *alert.Manager, token string, chatID int64, longPollerTimeout time.Duration) (*Service, error) {
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
		ID: chatID,
	}
	return &Service{
		alerts:   alerts,
		bot:      bot,
		chatID:   telebot.ChatID(chatID),
		chat:     &chat,
		handlers: make(map[string][]handler),
	}, nil
}

func NewBoardwhiteServiceFromConfig(cfg config.Config, alerts *alert.Manager) (*Service, error) {
	tgService, err := NewService(alerts, cfg.Tg.Key, cfg.Boardwhite.ChatID, cfg.Tg.LongPollerTimeout)
	if err != nil {
		return nil, fmt.Errorf("new tg client: %w", err)
	}

	return tgService, nil
}

func NewAdminClientFromConfig(cfg config.Config) (AdminClient, error) {
	// client doesn't need alerts
	tgService, err := NewService(nil, cfg.Tg.Key, cfg.Tg.AdminChatID, cfg.Tg.LongPollerTimeout)
	if err != nil {
		return nil, fmt.Errorf("new tg client: %w", err)
	}

	return tgService, nil
}

func SetReactionFor(c Client, messageID int) func(Reaction) error {
	return func(reaction Reaction) error {
		if err := c.SetReaction(messageID, reaction, false); err != nil {
			return fmt.Errorf("set reaction %v: %w", reaction, err)
		}

		return nil
	}
}

type handler struct {
	name string
	f    tele.HandlerFunc
}

func (s *Service) RegisterHandler(endpoint string, name string, f tele.HandlerFunc) {
	slog.Info("registered handler", slog.String("endpoint", endpoint), slog.String("name", name))
	s.handlers[endpoint] = append(s.handlers[endpoint], handler{
		name: name,
		f:    f,
	})
}

func wrapErrors(alerts *alert.Manager, h handler) func(tele.Context) {
	return func(c tele.Context) {
		defer func() {
			if err := recover(); err != nil {
				alerts.Errorf("panic in tg handler %s: %s", h.name, fmt.Sprintf("%+v", err))
			}
		}()
		err := h.f(c)
		if err != nil {
			alerts.Errorxf(err, "err in tg handler %s", h.name)
		}
	}
}

func (s *Service) Start() {
	for endpoint, handlers := range s.handlers {
		h := func(tc tele.Context) error {
			for _, h := range handlers {
				wrapErrors(s.alerts, h)(tc)
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

func (s *Service) NewUpdateContext(u tele.Update) tele.Context {
	return s.bot.NewContext(u)
}

var mdSpecialChars = []rune{'_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!'}

func EscapeMD(text string) string {
	var result strings.Builder
	result.Grow(len(text))
	for _, char := range text {
		if slices.Contains(mdSpecialChars, char) {
			result.WriteRune('\\')
		}
		result.WriteRune(char)
	}
	return result.String()
}

func (s *Service) BotID() int64 {
	return s.bot.Me.ID
}

func (s *Service) SendMonospace(threadID int, text string) (int, error) {
	message, err := s.bot.Send(s.chatID, text, &tele.SendOptions{
		ThreadID: threadID,
		Entities: []tele.MessageEntity{
			{
				Type:   tele.EntityCode,
				Offset: 0, // TODO: this technically should be in utf-16 code points
				Length: len(text),
			},
		},
	})
	if err != nil {
		return 0, fmt.Errorf("send text %q: %w", text, err)
	}

	return message.ID, nil
}

// SendMarkdownV2 sends a message according to a markdownV2 formatting style
// https://core.telegram.org/bots/api#markdownv2-style
//
// *bold \*text*
// _italic \*text_
// __underline__
// ~strikethrough~
// ||spoiler||
// *bold _italic bold ~italic bold strikethrough ||italic bold strikethrough spoiler||~ __underline italic bold___ bold*
// [inline URL](http://www.example.com/)
// [inline mention of a user](tg://user?id=123456789)
// ![👍](tg://emoji?id=5368324170671202286)
// `inline fixed-width code`
// ```
// pre-formatted fixed-width code block
// ```
// ```python
// pre-formatted fixed-width code block written in the Python programming language
// ```
// >Block quotation started
// >Block quotation continued
// >The last line of the block quotation**
// >The second block quotation started right after the previous\r
// >The third block quotation started right after the previous
func (s *Service) SendMarkdownV2(threadID int, text string) (int, error) {
	message, err := s.bot.Send(s.chatID, text, &tele.SendOptions{
		ThreadID:  threadID,
		ParseMode: tele.ModeMarkdownV2,
	})
	if err != nil {
		return 0, fmt.Errorf("send markdownV2 %q: %w", text, err)
	}

	return message.ID, nil
}

func (s *Service) SendText(threadID int, text string) (int, error) {
	message, err := s.bot.Send(s.chatID, text, &tele.SendOptions{
		ThreadID: threadID,
	})
	if err != nil {
		return 0, fmt.Errorf("send text %q: %w", text, err)
	}

	return message.ID, nil
}

func (s *Service) SendSpoilerLink(threadID int, header, link string) (int, error) {
	payload := fmt.Sprintf("%s\n%s", header, link)
	message, err := s.bot.Send(s.chatID, payload, &tele.SendOptions{
		ThreadID:              threadID,
		DisableWebPagePreview: true,
		Entities: []tele.MessageEntity{
			{
				Type:   tele.EntitySpoiler,
				Offset: len(header) + 1, // TODO: this technically should be in utf-16 code points
				Length: len(link),
			},
		},
	})
	if err != nil {
		return 0, fmt.Errorf("send spoiler link %q %q: %w", header, link, err)
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
		return 0, fmt.Errorf("send sticker %q: %w", stickerID, err)
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
		return 0, fmt.Errorf("reply with sticker %q: %w", stickerID, err)
	}

	return message.ID, nil
}

func (s *Service) ReplyWithSpoilerPhoto(messageID int, caption, name, mime string, reader io.ReadSeeker) (int, error) {
	var message *tele.Message
	var err error

	photo := tele.Photo{
		File:    tele.FromReader(reader),
		Caption: caption,
	}
	opts := tele.SendOptions{
		ReplyTo: &tele.Message{
			ID: messageID,
		},
		HasSpoiler: true,
	}
	message, err = s.bot.Send(s.chatID, &photo, &opts)
	if err != nil && strings.Contains(err.Error(), "PHOTO_INVALID_DIMENSIONS") {
		if _, err := reader.Seek(0, io.SeekStart); err != nil {
			return 0, fmt.Errorf("err seek at start: %w", err)
		}
		doc := tele.Document{
			File:     tele.FromReader(reader),
			FileName: name,
			MIME:     mime,
		}
		message, err = s.bot.Send(s.chatID, &doc, &opts)
	}
	if err != nil {
		return 0, fmt.Errorf("reply with spoiler photo: %w", err)
	}

	return message.ID, nil
}

func (s *Service) ReplyWithDocument(messageID int, name, mime string, reader io.ReadSeeker) (int, error) {
	opts := tele.SendOptions{
		ReplyTo: &tele.Message{
			ID: messageID,
		},
	}
	doc := tele.Document{
		File:     tele.FromReader(reader),
		FileName: name,
		MIME:     mime,
	}
	message, err := s.bot.Send(s.chatID, &doc, &opts)
	if err != nil {
		return 0, fmt.Errorf("reply with document: %w", err)
	}

	return message.ID, nil
}

func (s *Service) ReplyWithText(messageID int, text string) (int, error) {
	opts := tele.SendOptions{
		ReplyTo: &tele.Message{
			ID: messageID,
		},
	}
	message, err := s.bot.Send(s.chatID, text, &opts)
	if err != nil {
		return 0, fmt.Errorf("reply with text: %w", err)
	}

	return message.ID, nil
}

func (s *Service) EditMessageText(messageID int, newText string) error {
	msg := tele.StoredMessage{
		MessageID: strconv.Itoa(messageID),
		ChatID:    s.chat.ID,
	}
	if _, err := s.bot.Edit(msg, newText); err != nil {
		return fmt.Errorf("edit message text: %w", err)
	}

	return nil
}

func (s *Service) Pin(id int) error {
	msg := tele.StoredMessage{
		MessageID: strconv.Itoa(id),
		ChatID:    s.chat.ID,
	}
	if err := s.bot.Pin(msg, tele.Silent); err != nil {
		return fmt.Errorf("pin msg %v: %w", id, err)
	}

	return nil
}

func (s *Service) Unpin(id int) error {
	if err := s.bot.Unpin(s.chat, id); err != nil {
		return fmt.Errorf("unpin msg %v: %w", id, err)
	}

	return nil
}

type setMessageReactionReq struct {
	ChatID    tele.ChatID `json:"chat_id"`
	MessageID int         `json:"message_id"`
	Reactions []Reaction  `json:"reaction"`
	IsBig     bool        `json:"is_big,omitempty"`
}

func (s *Service) SetReaction(messageID int, reaction Reaction, isBig bool) error {
	req := setMessageReactionReq{
		ChatID:    s.chatID,
		MessageID: messageID,
		// currently, as non-premium users, bots can set up to one reaction per message
		Reactions: []Reaction{reaction},
		IsBig:     isBig,
	}
	_, err := s.bot.Raw("setMessageReaction", req)
	if err != nil {
		return fmt.Errorf("set reaction %v: %w", reaction, err)
	}

	return nil
}

func (s *Service) Delete(id int) error {
	msg := tele.StoredMessage{
		MessageID: strconv.Itoa(id),
		ChatID:    s.chat.ID,
	}
	err := s.bot.Delete(msg)
	if err != nil {
		return fmt.Errorf("delete msg %v: %w", id, err)
	}

	return nil
}

func (s *Service) SendAlert(msg string) error {
	chunkLen := 4096
	for i := 0; i < len(msg); i += chunkLen {
		chunk := msg[i:min(i+chunkLen, len(msg))]
		if _, err := s.SendMonospace(0, chunk); err != nil {
			return fmt.Errorf("send alert chunk %q: %w", chunk, err)
		}
	}

	return nil
}

func BuildMentionMarkdownV2(user tele.User) (name string) {
	if len(user.Username) > 0 {
		name = "@" + EscapeMD(user.Username)
	} else {
		fullName := iterx.JoinNonEmpty(" ", user.FirstName, user.LastName)
		name = fmt.Sprintf("[%s](tg://user?id=%d)", EscapeMD(fullName), user.ID)
	}
	return name
}
