package boardwhite

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/iter"
	"github.com/boar-d-white-foundation/drone/neetcode"
	"github.com/boar-d-white-foundation/drone/tg"
	tele "gopkg.in/telebot.v3"
)

func (s *Service) PublishNCDaily(ctx context.Context) error {
	dayIndex := int(time.Since(s.cfg.DailyNCStartDate).Hours()/24) % neetcode.QuestionsTotalCount
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

	return s.database.Do(ctx, func(tx db.Tx) error {
		messageID, err := s.publish(tx, header, link.String(), stickerID, keyNCPinnedMessages)
		if err != nil {
			return fmt.Errorf("publish: %w", err)
		}

		msgIDToDayIdx, err := db.GetJsonDefault(tx, keyNCPinnedToDayIdx, make(map[int]int64))
		if err != nil {
			return fmt.Errorf("get msgIDToDayIdx: %w", err)
		}

		msgIDToDayIdx[messageID] = int64(dayIndex)
		err = db.SetJson(tx, keyNCPinnedToDayIdx, msgIDToDayIdx)
		if err != nil {
			return fmt.Errorf("set msgIDToDayIdx: %w", err)
		}

		return nil
	})
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

		stats, err := db.GetJsonDefault[stats](tx, keyNCStats, newStats())
		if err != nil {
			return fmt.Errorf("get nc stats: %w", err)
		}

		msgIDToDayIdx, err := db.GetJsonDefault(tx, keyNCPinnedToDayIdx, make(map[int]int64))
		if err != nil {
			return fmt.Errorf("get msgIDToDayIdx: %w", err)
		}

		currentDayIdx, ok := msgIDToDayIdx[msg.ReplyTo.ID]
		if !ok {
			return setClown()
		}
		key := solutionKey{
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

func (s *Service) PublishNCRating(ctx context.Context) error {
	return s.database.Do(ctx, func(tx db.Tx) error {
		stats, err := db.GetJson[stats](tx, keyNCStats)
		switch {
		case err == nil:
		case errors.Is(err, db.ErrKeyNotFound):
			return nil
		default:
			return fmt.Errorf("get nc stats: %w", err)
		}

		report := buildRating(stats).toMarkdownV2("Neetcode leaderboard:")
		_, err = s.telegram.SendMarkdownV2(s.cfg.LeetcodeThreadID, report)
		if err != nil {
			return fmt.Errorf("send report: %w", err)
		}

		return nil
	})
}
