package boardwhite

import (
	"context"
	"fmt"

	"github.com/boar-d-white-foundation/drone/tg"
	tele "gopkg.in/telebot.v3"
)

func (s *Service) RegisterHandlers(ctx context.Context, registry tg.HandlerRegistry) {
	lcStatsHandler := s.makeStatsHandler(
		keyLCPinnedMessages,
		keyLCPinnedToStatsDayInfo,
		keyLCStats,
		ratingOpts{},
	)
	lcChickensStatsHandler := s.makeStatsHandler(
		keyLCChickensPinnedMessages,
		keyLCChickensPinnedToStatsDayInfo,
		keyLCChickensStats,
		ratingOpts{noComplexityEstimations: true},
	)
	ncStatsHandler := s.makeStatsHandler(
		keyNCPinnedMessages,
		keyNCPinnedToStatsDayInfo,
		keyNCStats,
		ratingOpts{},
	)

	registry.RegisterHandler(tele.OnText, "OnLeetCodeUpdateText", withContext(ctx, lcStatsHandler))
	registry.RegisterHandler(tele.OnPhoto, "OnLeetCodeUpdatePhoto", withContext(ctx, lcStatsHandler))
	registry.RegisterHandler(tele.OnText, "OnLeetCodeChickensUpdateText", withContext(ctx, lcChickensStatsHandler))
	registry.RegisterHandler(tele.OnPhoto, "OnLeetCodeChickensUpdatePhoto", withContext(ctx, lcChickensStatsHandler))
	registry.RegisterHandler(tele.OnText, "OnNeetCodeUpdateText", withContext(ctx, ncStatsHandler))
	registry.RegisterHandler(tele.OnPhoto, "OnNeetCodeUpdatePhoto", withContext(ctx, ncStatsHandler))
	registry.RegisterHandler(tele.OnText, "OnMock", withContext(ctx, s.OnMock))
	registry.RegisterHandler(tele.OnPinned, "OnBotPinned", withContext(ctx, s.OnBotPinned))
	registry.RegisterHandler(tele.OnUserJoined, "OnGreetJoinedUser", withContext(ctx, s.OnGreetJoinedUser))
	registry.RegisterHandler(tele.OnText, "OnGenerateVCPdf", withContext(ctx, s.OnGenerateVCPdf))
	registry.RegisterHandler(tele.OnText, "OnPostTwitterEmbed", withContext(ctx, s.OnPostTwitterEmbed))
	registry.RegisterHandler(tele.OnText, "OnOborona", withContext(ctx, s.OnOborona))
	registry.RegisterHandler(tele.OnText, "OnUpdateOkr", withContext(ctx, s.OnUpdateOkr))
	registry.RegisterHandler(tele.OnText, "OnRemoveOkr", withContext(ctx, s.OnRemoveOkr))
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
