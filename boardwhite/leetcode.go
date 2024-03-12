package boardwhite

import (
	"context"
	"fmt"

	"github.com/frosthamster/drone/leetcode"
)

const (
	defaultDailyHeader = "LeetCode Daily Question"

	keyLeetcodePinnedMessage = "leetcode:pinned_message"
)

func (s *Service) PublishLCDaily(ctx context.Context) error {
	link, err := leetcode.GetDailyLink(ctx)
	if err != nil {
		return fmt.Errorf("get link: %w", err)
	}

	key := []byte(keyLeetcodePinnedMessage)

	return s.publish(defaultDailyHeader, link, s.dailyLCStickerID, key)
}
