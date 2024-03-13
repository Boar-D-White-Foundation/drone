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
	keyNeetcodePinnedMessage = "neetcode:pinned_message"
)

func (s *Service) PublishNCDaily(ctx context.Context) error {
	groups, err := neetcode.Groups()
	if err != nil {
		return fmt.Errorf("read groups: %w", err)
	}

	totalQuestions := 0
	for _, g := range groups {
		totalQuestions += len(g.Questions)
	}
	dayIndex := int(time.Now().Sub(s.dailyNCStartDate).Hours()/24) % totalQuestions

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

	header := fmt.Sprintf("NeetCode: %s [%d / %d]", group.Name, dayIndex+1, totalQuestions)

	var link strings.Builder
	link.WriteString(question.LCLink)
	if len(question.FreeLink) > 0 {
		link.WriteString("\n")
		link.WriteString(question.FreeLink)
	}

	var stickerID string
	if group.Name == "1-D DP" || group.Name == "2-D DP" {
		stickerID = s.dpStickerID
	} else {
		stickerID, err = iter.PickRandom(s.dailyStickersIDs)
		if err != nil {
			return fmt.Errorf("get sticker: %w", err)
		}
	}

	key := []byte(keyNeetcodePinnedMessage)
	return s.publish(header, link.String(), stickerID, key)
}
