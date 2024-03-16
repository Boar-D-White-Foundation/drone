package boardwhite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/iter"
	"github.com/boar-d-white-foundation/drone/neetcode"
	"github.com/boar-d-white-foundation/drone/tg"
	tele "gopkg.in/telebot.v3"
)

const (
	keyNeetCodePinnedMessage = "boardwhite:neetcode:pinned_message"
)

func (s *Service) newNeetCodeCounter(ctx context.Context) tele.HandlerFunc {

	return func(c tele.Context) error {
		message := c.Message()
		// ignore not replies or replies not to bot early
		if message.ReplyTo == nil || message.ReplyTo.OriginalSender != c.Bot().Me {
			return nil
		}
		err := s.database.Do(ctx, func(tx db.Tx) error {
			pinnedMessageId, err := db.GetJson[int](tx, keyNeetCodePinnedMessage)
			if err != nil {
				return err
			}
			// not a reply to our current pin
			if message.ReplyTo.ID != pinnedMessageId {
				return nil
			}

			correctFormat := true
			if len(message.Text) > 0 {
				correctFormat = strings.Contains(message.Text, "leetcode.com")
			}
			if correctFormat {
				return s.telegram.SetMessageReaction(message, tg.ReactionThumbsUp, true)
			}
			return s.telegram.SetMessageReaction(message, tg.ReactionClown, true)
		})
		return err
	}
}

func (s *Service) PublishNCDaily(ctx context.Context) error {
	groups, err := neetcode.Groups()
	if err != nil {
		return fmt.Errorf("read groups: %w", err)
	}

	totalQuestions := 0
	for _, g := range groups {
		totalQuestions += len(g.Questions)
	}
	dayIndex := int(time.Now().Sub(s.DailyNCStartDate).Hours()/24) % totalQuestions

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
		stickerID = s.DpStickerID
	} else {
		stickerID, err = iter.PickRandom(s.DailyStickersIDs)
		if err != nil {
			return fmt.Errorf("get sticker: %w", err)
		}
	}

	return s.publish(ctx, header, link.String(), stickerID, keyNeetCodePinnedMessage)
}
