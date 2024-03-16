package boardwhite

import (
	"context"
	"errors"
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

type neetCodeCounter struct {
	ctx      context.Context
	database db.DB
}

func newNeetCodeCounter(ctx context.Context, db db.DB) *neetCodeCounter {
	return &neetCodeCounter{
		ctx:      ctx,
		database: db,
	}
}

func (n neetCodeCounter) Match(c tele.Context) bool {

	message := c.Message()

	// not a reply
	if message.ReplyTo == nil {
		return false
	}

	err := n.database.Do(n.ctx, func(tx db.Tx) error {
		pinnedMessageId, err := db.GetJson[int](tx, keyNeetCodePinnedMessage)

		if err != nil {
			return err
		}

		if message.ReplyTo.ID == pinnedMessageId {
			return nil
		}

		return errors.New("not a reply to the pinned message")

	})

	return err == nil
}

func (n neetCodeCounter) Handle(client *tg.Client, c tele.Context) error {
	message := c.Message()
	return client.SetMessageReaction(message, tg.ReactionThumbsUp, true)
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
