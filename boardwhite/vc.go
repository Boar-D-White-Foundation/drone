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

	cmdPrefix := "/pdf"
	if !(strings.HasPrefix(msg.Text, cmdPrefix) || strings.HasPrefix(msg.Caption, cmdPrefix)) {
		return nil
	}

	set := tg.SetReactionFor(s.telegram, msg.ID)
	link := s.getVcLink(msg)
	if link == "" {
		return set(tg.ReactionClown)
	}

	if s.mediaGenerator == nil {
		return set(tg.ReactionHeadExplode)
	}

	if err := set(tg.ReactionEyes); err != nil {
		return fmt.Errorf("set progress reaction for vc pdf generation: %w", err)
	}

	// TODO: move to dbq
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

func (s *Service) getVcLink(msg *tele.Message) string {
	options := make([]string, 0, 4)
	options = append(options, msg.Text, msg.Caption)
	if msg.ReplyTo != nil {
		options = append(options, msg.ReplyTo.Text, msg.ReplyTo.Caption)
	}

	for _, opt := range options {
		match := s.vcLinkRe.FindStringSubmatch(opt)
		if len(match) >= 2 {
			return match[1]
		}
	}

	return ""
}
