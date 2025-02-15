package boardwhite

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/boar-d-white-foundation/drone/tg"
	tele "gopkg.in/telebot.v3"
)

func (s *Service) OnGenerateVCPdf(ctx context.Context, c tele.Context) error {
	msg, chat := c.Message(), c.Chat()
	if msg == nil || chat == nil || chat.ID != s.cfg.ChatID {
		return nil
	}
	if !strings.HasPrefix(msg.Text, "/pdf") {
		return nil
	}

	set := tg.SetReactionFor(s.telegram, msg.ID)
	match := s.vcLinkRe.FindStringSubmatch(msg.Text)
	if len(match) < 2 {
		return set(tg.ReactionClown)
	}

	if err := set(tg.ReactionEyes); err != nil {
		return fmt.Errorf("set progress reaction for vc pdf generation: %w", err)
	}

	// TODO: move to dbq
	link := match[1]
	buf, err := s.mediaGenerator.GenerateVCPagePdf(ctx, link)
	if err != nil {
		if err := set(tg.ReactionHeadExplode); err != nil {
			s.alerts.Errorxf(err, "failed to set fail reaction for vc pdf generation")
		}
		return fmt.Errorf("generate vc pdf: %w", err)
	}

	_, err = s.telegram.ReplyWithDocument(msg.ID, "dump.pdf", "application/pdf", bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("reply with vc dump: %w", err)
	}

	return nil
}
