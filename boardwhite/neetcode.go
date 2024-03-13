package boardwhite

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	for _, g := range groups {
		if dayIndex < len(g.Questions) {
			group = g
			question = g.Questions[dayIndex]
			break
		}
		dayIndex -= len(g.Questions)
	}

	header := fmt.Sprintf("NeetCode: %s [%d / 150]", group.Name, dayIndex+1)

	var link strings.Builder
	link.WriteString(question.LCLink)
	if len(question.FreeLink) > 0 {
		link.WriteString("\n")
		link.WriteString(question.FreeLink)
	}

	key := []byte(keyNeetcodePinnedMessage)
	return s.publish(header, link.String(), s.dailyNCStickerID, key)
}
