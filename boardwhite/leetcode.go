package boardwhite

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/iter"
	"github.com/boar-d-white-foundation/drone/leetcode"
	"golang.org/x/exp/rand"
)

const (
	defaultDailyHeader        = "LeetCode Daily Question"
	defaultDailyChickenHeader = "LeetCode Daily Easy Question"
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
		lastDayInfo, err := s.getLastPublishedQuestionDayInfo(tx, keyLCPinnedToStatsDayInfo)
		if err != nil {
			return fmt.Errorf("get last published lc question: %w", err)
		}

		_, err = s.publishDaily(tx, publishDailyReq{
			dayIdx:          lastDayInfo.DayIdx + 1,
			threadID:        s.cfg.LeetcodeThreadID,
			header:          defaultDailyHeader,
			text:            dailyInfo.Link,
			stickerID:       stickerID,
			pinnedMsgsKey:   keyLCPinnedMessages,
			msgToDayInfoKey: keyLCPinnedToStatsDayInfo,
		})
		if err != nil {
			return fmt.Errorf("publish lc daily: %w", err)
		}

		return nil
	})
}

func (s *Service) PublishLCRating(ctx context.Context) error {
	return s.publishRating(
		ctx,
		30,
		"Leetcode leaderboard (last 30 questions):",
		s.cfg.LeetcodeThreadID,
		keyLCPinnedToStatsDayInfo,
		keyLCStats,
	)
}

type lcChickenQuestions struct {
	questions        []leetcode.Question
	shuffledPosition []int
}

func newLCChickenQuestions() (lcChickenQuestions, error) {
	questions, err := leetcode.Questions()
	if err != nil {
		return lcChickenQuestions{}, fmt.Errorf("get lc questions: %w", err)
	}

	easyQuestions := make([]leetcode.Question, 0)
	for _, q := range questions {
		if q.Difficulty == leetcode.DifficultyEasy && !q.IsPremium {
			easyQuestions = append(easyQuestions, q)
		}
	}

	// use pseudorandom deterministic order for fallback questions
	rnd := rand.New(rand.NewSource(315135236))
	shuffledPosition := make([]int, 0, len(easyQuestions))
	for i := range easyQuestions {
		shuffledPosition = append(shuffledPosition, i)
	}
	for i := range shuffledPosition {
		j := rnd.Intn(i + 1)
		shuffledPosition[i], shuffledPosition[j] = shuffledPosition[j], shuffledPosition[i]
	}

	result := lcChickenQuestions{
		questions:        easyQuestions,
		shuffledPosition: shuffledPosition,
	}
	return result, nil
}

func (cq *lcChickenQuestions) getNextQuestion(tx db.Tx) (leetcode.Question, error) {
	idx, err := db.GetJsonDefault[int](tx, keyLCChickensFallbackQuestionIdx, 0)
	if err != nil {
		return leetcode.Question{}, fmt.Errorf("get idx: %w", err)
	}

	question := cq.questions[cq.shuffledPosition[idx%len(cq.questions)]]
	idx++
	if err := db.SetJson(tx, keyLCChickensFallbackQuestionIdx, idx); err != nil {
		return leetcode.Question{}, fmt.Errorf("set keyLCChickensFallbackQuestionIdx: %w", err)
	}

	return question, nil
}

func (s *Service) PublishLCChickensDaily(ctx context.Context) error {
	dailyInfo, err := leetcode.GetDailyInfo(ctx)
	if err != nil {
		return fmt.Errorf("get link: %w", err)
	}

	stickerID, err := iter.PickRandom(s.cfg.DailyChickensStickerIDs)
	if err != nil {
		return fmt.Errorf("get sticker: %w", err)
	}

	return s.database.Do(ctx, func(tx db.Tx) error {
		lastDayInfo, err := s.getLastPublishedQuestionDayInfo(tx, keyLCChickensPinnedToStatsDayInfo)
		if err != nil {
			return fmt.Errorf("get last published lc question: %w", err)
		}

		link, err := s.selectLCChickensDailyLink(tx, dailyInfo)
		if err != nil {
			return fmt.Errorf("select link: %w", err)
		}

		_, err = s.publishDaily(tx, publishDailyReq{
			dayIdx:          lastDayInfo.DayIdx + 1,
			threadID:        s.cfg.LeetcodeChickensThreadID,
			header:          defaultDailyChickenHeader,
			text:            link,
			stickerID:       stickerID,
			pinnedMsgsKey:   keyLCChickensPinnedMessages,
			msgToDayInfoKey: keyLCChickensPinnedToStatsDayInfo,
		})
		if err != nil {
			return fmt.Errorf("publish lc checkens daily: %w", err)
		}

		return nil
	})
}

func (s *Service) selectLCChickensDailyLink(
	tx db.Tx,
	dailyInfo leetcode.DailyInfo,
) (string, error) {
	if dailyInfo.Difficulty == leetcode.DifficultyEasy {
		slog.Info("daily is easy, selecting it")
		return dailyInfo.Link, nil
	}

	slog.Info("daily is not easy, selecting fallback")
	question, err := s.lcChickenQuestions.getNextQuestion(tx)
	if err != nil {
		return "", fmt.Errorf("get fallback question: %w", err)
	}

	return question.Link, nil
}

func (s *Service) PublishLCChickensRating(ctx context.Context) error {
	return s.publishRating(
		ctx,
		30,
		"Leetcode easy leaderboard (last 30 questions):",
		s.cfg.LeetcodeChickensThreadID,
		keyLCChickensPinnedToStatsDayInfo,
		keyLCChickensStats,
	)
}
