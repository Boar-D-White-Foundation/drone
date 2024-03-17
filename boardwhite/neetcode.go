package boardwhite

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/iter"
	"github.com/boar-d-white-foundation/drone/neetcode"
	"github.com/boar-d-white-foundation/drone/tg"
	tele "gopkg.in/telebot.v3"
)

func (s *Service) getNCDayIdx() int {
	return int(time.Since(s.cfg.DailyNCStartDate).Hours()/24) % neetcode.QuestionsTotalCount
}

func (s *Service) PublishNCDaily(ctx context.Context) error {
	dayIndex := s.getNCDayIdx()
	groups, err := neetcode.Groups()
	if err != nil {
		return fmt.Errorf("read groups: %w", err)
	}

	var group neetcode.Group
	var question neetcode.Question
	idx := dayIndex
	for _, g := range groups {
		if idx < len(g.Questions) {
			group = g
			question = g.Questions[idx]
			break
		}
		idx -= len(g.Questions)
	}

	header := fmt.Sprintf("NeetCode: %s [%d / %d]", group.Name, dayIndex+1, neetcode.QuestionsTotalCount)

	var link strings.Builder
	link.WriteString(question.LCLink)
	if len(question.FreeLink) > 0 {
		link.WriteString("\n")
		link.WriteString(question.FreeLink)
	}

	var stickerID string
	if group.Name == "1-D DP" || group.Name == "2-D DP" {
		stickerID = s.cfg.DpStickerID
	} else {
		stickerID, err = iter.PickRandom(s.cfg.DailyStickersIDs)
		if err != nil {
			return fmt.Errorf("get sticker: %w", err)
		}
	}

	return s.publish(ctx, header, link.String(), stickerID, keyNCPinnedMessages)
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
