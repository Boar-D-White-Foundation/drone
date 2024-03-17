package boardwhite

import (
	"context"

	"github.com/boar-d-white-foundation/drone/tg"
	tele "gopkg.in/telebot.v3"
)

func (s *Service) RegisterHandlers(ctx context.Context, tgService *tg.Service) {
	tgService.RegisterHandler(tele.OnText, "OnNeetCodeUpdateText", withContext(ctx, s.OnNeetCodeUpdate))
	tgService.RegisterHandler(tele.OnPhoto, "OnNeetCodeUpdatePhoto", withContext(ctx, s.OnNeetCodeUpdate))
	if s.cfg.MockEgor.Enabled {
		tgService.RegisterHandler(tele.OnText, "OnMockEgor", withContext(ctx, s.OnMockEgor))
	}
}

func withContext(ctx context.Context, f func(context.Context, tele.Context) error) tele.HandlerFunc {
	return func(c tele.Context) error {
		return f(ctx, c)
	}
}
