package boardwhite

import (
	"context"
	"fmt"

	"github.com/boar-d-white-foundation/drone/tg"
	tele "gopkg.in/telebot.v3"
)

func (s *Service) RegisterHandlers(ctx context.Context, tgService *tg.Service) {
	tgService.RegisterHandler(tele.OnText, "OnNeetCodeUpdateText", withContext(ctx, s.OnNeetCodeUpdate))
	tgService.RegisterHandler(tele.OnPhoto, "OnNeetCodeUpdatePhoto", withContext(ctx, s.OnNeetCodeUpdate))
	tgService.RegisterHandler(tele.OnText, "OnMock", withContext(ctx, s.OnMock))
	tgService.RegisterHandler(tele.OnPinned, "OnBotPinned", withContext(ctx, s.OnBotPinned))
}

func withContext(ctx context.Context, f func(context.Context, tele.Context) error) tele.HandlerFunc {
	return func(c tele.Context) error {
		return f(ctx, c)
	}
}

func (s *Service) OnBotPinned(ctx context.Context, c tele.Context) error {
	msg := c.Message()
	if msg == nil || msg.PinnedMessage == nil || msg.Sender == nil || msg.Sender.ID != s.telegram.BotID() {
		return nil
	}

	err := s.telegram.Delete(msg.ID)
	if err != nil {
		return fmt.Errorf("delete pinned message notification: %w", err)
	}

	return nil
}
