package boardwhite

import (
	"context"
	"fmt"
	"time"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/iterx"
	tele "gopkg.in/telebot.v3"
)

func (s *Service) generateOborona() (string, error) {
	picks := make([]any, 0, len(s.cfg.OboronaWords))
	for _, words := range s.cfg.OboronaWords {
		pick, err := iterx.PickRandom(words)
		if err != nil {
			return "", fmt.Errorf("pick random word: %w", err)
		}

		picks = append(picks, pick)
	}

	return fmt.Sprintf(s.cfg.OboronaTemplate, picks...), nil
}

func (s *Service) OnOborona(ctx context.Context, c tele.Context) error {
	msg, chat := c.Message(), c.Chat()
	if msg == nil || chat == nil || chat.ID != s.cfg.ChatID {
		return nil
	}

	return s.database.Do(ctx, func(tx db.Tx) error {
		generatedAt, err := db.GetJsonDefault[time.Time](tx, keyOboronaLastGeneratedAt, time.Time{})
		if err != nil {
			return fmt.Errorf("get oborona generated at: %w", err)
		}

		if time.Since(generatedAt) < s.cfg.OboronaPeriod {
			return nil
		}

		oborona, err := s.generateOborona()
		if err != nil {
			return fmt.Errorf("generate oborona: %w", err)
		}

		_, err = s.telegram.ReplyWithText(msg.ID, oborona)
		if err != nil {
			return fmt.Errorf("reply with oborona: %w", err)
		}

		if err := db.SetJson(tx, keyOboronaLastGeneratedAt, time.Now()); err != nil {
			return fmt.Errorf("set oborona generated at: %w", err)
		}

		return nil
	})
}
