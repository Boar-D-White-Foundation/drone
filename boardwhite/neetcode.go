package boardwhite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/boar-d-white-foundation/drone/iter"
	"github.com/boar-d-white-foundation/drone/neetcode"
)

const (
	ncTotalDays = 150
)

func (s *Service) getNCDayIdx() int {
	return int(time.Since(s.cfg.DailyNCStartDate).Hours()/24) % ncTotalDays
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

	header := fmt.Sprintf("NeetCode: %s [%d / %d]", group.Name, dayIndex+1, ncTotalDays)

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
