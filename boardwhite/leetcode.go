package boardwhite

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/iter"
	"github.com/boar-d-white-foundation/drone/leetcode"
)

const (
	defaultDailyHeader = "LeetCode Daily Question"
)

func (s *Service) PublishLCDaily(ctx context.Context) error {
	dailyInfo, err := leetcode.GetDailyInfo(ctx)
	if err != nil {
		return fmt.Errorf("get link: %w", err)
	}

	stickerID, err := iter.PickRandom(s.cfg.DailyStickersIDs)
	if err != nil {
		return fmt.Errorf("get sticker: %w", err)
	}

	return s.database.Do(ctx, func(tx db.Tx) error {
		_, err := s.publish(
			tx,
			s.cfg.LeetcodeThreadID,
			defaultDailyHeader,
			dailyInfo.Link,
			stickerID,
			keyLCPinnedMessages,
		)
		return err
	})
}

func (s *Service) PublishLCChickensDaily(ctx context.Context) error {
	dailyInfo, err := leetcode.GetDailyInfo(ctx)
	if err != nil {
		return fmt.Errorf("get link: %w", err)
	}

	if dailyInfo.Difficulty != leetcode.DifficultyEasy {
		slog.Info("daily is not easy, skipping")
		return nil
	}

	stickerID, err := iter.PickRandom(s.cfg.DailyChickensStickerIDs)
	if err != nil {
		return fmt.Errorf("get sticker: %w", err)
	}

	return s.database.Do(ctx, func(tx db.Tx) error {
		_, err := s.publish(
			tx,
			s.cfg.LeetcodeChickensThreadID,
			defaultDailyHeader,
			dailyInfo.Link,
			stickerID,
			keyLCChickensPinnedMessages,
		)
		return err
	})
}
