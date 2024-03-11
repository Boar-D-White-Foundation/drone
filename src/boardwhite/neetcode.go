package boardwhite

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/frosthamster/drone/src/neetcode"
)

const (
	neetcodeDailyHeader = "NeetCode Daily Question"

	keyNeetcodePinnedMessage = "neetcode:pinned_message"
)

func (s *Service) PublishNCDaily(ctx context.Context) error {
	weekday := time.Now().Weekday()
	difficulty := weekdayToDifficulty(weekday)

	qs, err := neetcode.QuestionsByDifficulty(difficulty)
	if err != nil {
		return fmt.Errorf("read questions: %w", err)
	}

	idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(qs))))
	if err != nil {
		return fmt.Errorf("generate random: %w", err)
	}
	q := qs[idx.Int64()]

	link := fmt.Sprintf("%s\n%s", q.LeetcodeLink(), q.LeetcodeCaLink())

	key := []byte(keyNeetcodePinnedMessage)

	return s.publish(neetcodeDailyHeader, link, key)
}

func weekdayToDifficulty(wd time.Weekday) string {
	switch wd {
	case time.Monday:
		return "easy"
	case time.Thursday:
		return "hard"
	default:
		return "medium"
	}
}
