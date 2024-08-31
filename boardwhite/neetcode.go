package boardwhite

import (
	"context"
	"fmt"
	"strings"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/iterx"
	"github.com/boar-d-white-foundation/drone/neetcode"
)

func (s *Service) PublishNCDaily(ctx context.Context) error {
	return s.database.Do(ctx, func(tx db.Tx) error {
		lastDayInfo, err := s.getLastPublishedQuestionDayInfo(tx, keyNCPinnedToStatsDayInfo)
		if err != nil {
			return fmt.Errorf("get last published nc question: %w", err)
		}

		dayIdx := int((lastDayInfo.DayIdx + 1) % neetcode.QuestionsTotalCount)
		groups, err := neetcode.Groups()
		if err != nil {
			return fmt.Errorf("read groups: %w", err)
		}

		var group neetcode.Group
		var question neetcode.Question
		idx := dayIdx
		for _, g := range groups {
			if idx < len(g.Questions) {
				group = g
				question = g.Questions[idx]
				break
			}
			idx -= len(g.Questions)
		}

		header := fmt.Sprintf("NeetCode: %s [%d / %d]", group.Name, dayIdx+1, neetcode.QuestionsTotalCount)

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
			stickerID, err = iterx.PickRandom(s.cfg.DailyStickersIDs)
			if err != nil {
				return fmt.Errorf("get sticker: %w", err)
			}
		}

		_, err = s.publishDaily(tx, publishDailyReq{
			dayIdx:          lastDayInfo.DayIdx + 1,
			threadID:        s.cfg.LeetcodeThreadID,
			header:          header,
			text:            link.String(),
			stickerID:       stickerID,
			pinnedMsgsKey:   keyNCPinnedMessages,
			msgToDayInfoKey: keyNCPinnedToStatsDayInfo,
		})
		if err != nil {
			return fmt.Errorf("publish nc daily: %w", err)
		}

		return nil
	})
}

func (s *Service) PublishNCRating(ctx context.Context) error {
	return s.publishRating(
		ctx,
		35,
		"Neetcode leaderboard (last 35 questions):",
		s.cfg.LeetcodeThreadID,
		keyNCPinnedToStatsDayInfo,
		keyNCStats,
	)
}
