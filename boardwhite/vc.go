package boardwhite

import (
	"bytes"
	"context"
	"fmt"

	tele "gopkg.in/telebot.v3"
)

func (s *Service) OnGenerateVCPdf(ctx context.Context, c tele.Context) error {
	msg, chat := c.Message(), c.Chat()
	if msg == nil || chat == nil || chat.ID != s.cfg.ChatID {
		return nil
	}

	match := s.vcLinkRe.FindStringSubmatch(msg.Text)
	if len(match) < 2 {
		return nil
	}

	buf, err := s.mediaGenerator.GenerateVCPagePdf(ctx, match[1])
	if err != nil {
		return fmt.Errorf("generate vc pdf: %w", err)
	}

	_, err = s.telegram.ReplyWithDocument(msg.ID, "dump.pdf", "application/pdf", bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("reply with vc dump: %w", err)
	}

	return nil
}
