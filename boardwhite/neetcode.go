package boardwhite

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/boar-d-white-foundation/drone/iter"
	"github.com/boar-d-white-foundation/drone/neetcode"
	"github.com/dgraph-io/badger/v4"
	tele "gopkg.in/telebot.v3"
)

const (
	keyNeetCodePinnedMessage = "boardwhite:neetcode:pinned_message"
)

type NeetCodeCounter struct {
	db *badger.DB
}

func (n NeetCodeCounter) Match(c tele.Context) bool {

	message := c.Message()

	// not a reply
	if message.ReplyTo == nil {
		return false
	}

	err := n.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(keyNeetcodePinnedMessage))
		// failed to get the key
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			pinnedId, _ := strconv.Atoi(string(val))
			if message.ReplyTo.ID == pinnedId {
				return nil
			}
			return errors.New("not a reply to the pinned message")
		})

	})
	return err == nil
}

func (n NeetCodeCounter) Handle(c tele.Context) error {

	// do something with the message
	return nil
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

	return s.publish(ctx, header, link.String(), stickerID, keyNeetCodePinnedMessage)
}
