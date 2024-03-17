package boardwhite

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/tg"
	tele "gopkg.in/telebot.v3"
)

const (
	byegor = "byegor"
)

var (
	lcSubmissionRe = regexp.MustCompile(`^https://leetcode.com/.+/submissions/[^/]+/?$`)
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

type ncSolutionKey struct {
	DayIdx int   `json:"day_idx"`
	UserID int64 `json:"user_id"`
}

func (s *ncSolutionKey) UnmarshalText(data []byte) error {
	parts := strings.Split(string(data), "|")
	if len(parts) != 2 {
		return errors.New("invalid parts")
	}
	dayIdx, err := strconv.Atoi(parts[0])
	if err != nil {
		return err
	}
	userID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return err
	}
	s.DayIdx = dayIdx
	s.UserID = userID
	return nil
}

func (s ncSolutionKey) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%d|%d", s.DayIdx, s.UserID)), nil
}

type solution struct {
	Update tele.Update `json:"update"`
}

type ncStats struct {
	Solutions map[ncSolutionKey]solution `json:"solutions,omitempty"`
}

func newNCStats() ncStats {
	return ncStats{
		Solutions: make(map[ncSolutionKey]solution),
	}
}

func (s *Service) OnNeetCodeUpdate(ctx context.Context, c tele.Context) error {
	update, msg, sender := c.Update(), c.Message(), c.Sender()
	if msg == nil || sender == nil {
		return nil
	}
	if msg.ReplyTo == nil || msg.ReplyTo.Sender.ID != c.Bot().Me.ID {
		return nil
	}

	setClown := func() error {
		return s.telegram.SetReaction(msg.ID, tg.ReactionClown, false)
	}
	return s.database.Do(ctx, func(tx db.Tx) error {
		pinnedIDs, err := db.GetJsonDefault[[]int](tx, keyNCPinnedMessages, nil)
		if err != nil {
			return fmt.Errorf("get pinnedIDs: %w", err)
		}
		if !slices.Contains(pinnedIDs, msg.ReplyTo.ID) {
			return nil
		}
		if msg.ReplyTo.ID != pinnedIDs[len(pinnedIDs)-1] {
			return setClown()
		}

		switch {
		case len(msg.Text) > 0:
			if !lcSubmissionRe.MatchString(msg.Text) {
				return setClown()
			}
		case msg.Photo != nil:
			if !msg.HasMediaSpoiler {
				return setClown()
			}
		default:
			return setClown()
		}

		stats, err := db.GetJsonDefault[ncStats](tx, keyNCStats, newNCStats())
		if err != nil {
			return fmt.Errorf("get solutions: %w", err)
		}

		currentDayIdx := s.getNCDayIdx() - 1
		if currentDayIdx < 0 {
			return setClown()
		}
		key := ncSolutionKey{
			DayIdx: currentDayIdx,
			UserID: sender.ID,
		}
		stats.Solutions[key] = solution{Update: update}
		err = db.SetJson(tx, keyNCStats, stats)
		if err != nil {
			return fmt.Errorf("set solutions: %w", err)
		}

		return s.telegram.SetReaction(msg.ID, tg.ReactionOk, false)
	})
}

func (s *Service) OnMockEgor(ctx context.Context, c tele.Context) error {
	msg, sender := c.Message(), c.Sender()
	if msg == nil || sender == nil {
		return nil
	}
	if sender.Username != byegor {
		return nil
	}

	key := keyLastByegorMockTime
	return s.database.Do(ctx, func(tx db.Tx) error {
		lastMock, err := db.GetJson[time.Time](tx, key)
		switch {
		case err == nil:
			if time.Since(lastMock) < s.cfg.MockEgor.Period {
				return nil
			}
		case errors.Is(err, db.ErrKeyNotFound):
			// still do it
		default:
			return fmt.Errorf("get lastMock: %w", err)
		}

		_, err = s.telegram.ReplyWithSticker(msg.ID, s.cfg.MockEgor.StickerID)
		if err != nil {
			return fmt.Errorf("reply with sticker: %w", err)
		}

		err = db.SetJson[time.Time](tx, key, time.Now())
		if err != nil {
			return fmt.Errorf("set lastMock: %w", err)
		}

		return nil
	})
}
