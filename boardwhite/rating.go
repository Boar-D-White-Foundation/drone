package boardwhite

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/iter"
	"github.com/boar-d-white-foundation/drone/tg"
	"golang.org/x/exp/slog"
	tele "gopkg.in/telebot.v3"
)

type solutionKey struct {
	DayIdx int64 `json:"day_idx"` // subsequent days must have values v, v+1
	UserID int64 `json:"user_id"`
}

func (k *solutionKey) UnmarshalText(data []byte) error {
	parts := strings.Split(string(data), "|")
	if len(parts) != 2 {
		return errors.New("invalid parts")
	}
	dayIdx, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return err
	}
	userID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return err
	}
	k.DayIdx = dayIdx
	k.UserID = userID
	return nil
}

func (k solutionKey) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%d|%d", k.DayIdx, k.UserID)), nil
}

type solution struct {
	Update tele.Update `json:"update"`
}

type statsDayInfo struct {
	DayIdx      int64     `json:"day_idx"`
	PublishedAt time.Time `json:"published_at"`
}

type stats struct {
	Solutions map[solutionKey]solution `json:"solutions"`
	DaysInfo  map[int64]statsDayInfo   `json:"days_info"`
}

func (s *Service) getLastPublishedQuestionDayInfo(tx db.Tx, msgToDayInfoKey string) (statsDayInfo, error) {
	result := statsDayInfo{DayIdx: -1}
	msgToDayInfo, err := db.GetJsonDefault(tx, msgToDayInfoKey, make(map[int]statsDayInfo))
	if err != nil {
		return statsDayInfo{}, fmt.Errorf("get msgToDayInfo: %w", err)
	}

	for _, dayInfo := range msgToDayInfo {
		if dayInfo.DayIdx > result.DayIdx {
			result = dayInfo
		}
	}

	return result, nil
}

func (s *Service) makeStatsHandler(
	pinnedMessagesKey string,
	msgToDayInfoKey string,
	statsKey string,
) func(context.Context, tele.Context) error {
	return func(ctx context.Context, c tele.Context) error {
		update, msg, sender := c.Update(), c.Message(), c.Sender()
		if msg == nil || sender == nil {
			return nil
		}
		if msg.ReplyTo == nil || msg.ReplyTo.Sender.ID != c.Bot().Me.ID {
			return nil
		}

		set := func(reaction tg.Reaction) error {
			return s.telegram.SetReaction(msg.ID, reaction, false)
		}
		return s.database.Do(ctx, func(tx db.Tx) error {
			pinnedIDs, err := db.GetJsonDefault[[]int](tx, pinnedMessagesKey, nil)
			if err != nil {
				return fmt.Errorf("get pinnedIDs: %w", err)
			}
			if !slices.Contains(pinnedIDs, msg.ReplyTo.ID) {
				return nil
			}

			switch {
			case len(msg.Text) > 0:
				match := lcSubmissionRe.FindStringSubmatch(msg.Text)
				if len(match) < 2 {
					return set(tg.ReactionClown)
				}
				if s.cfg.SnippetsGenerationEnabled {
					err := s.tasks.postCodeSnippet.Schedule(tx, 1, postCodeSnippetArgs{
						MessageID:    msg.ID,
						SubmissionID: match[1],
					})
					if err != nil {
						s.alerts.Errorxf(err, "err schedule post code snippet: %v", msg.Text)
					}
				}
			case msg.Photo != nil:
				if !msg.HasMediaSpoiler {
					return set(tg.ReactionClown)
				}
			default:
				return set(tg.ReactionClown)
			}

			if msg.ReplyTo.ID != pinnedIDs[len(pinnedIDs)-1] {
				return set(tg.ReactionMoai) // deadline miss
			}

			stats, err := db.GetJsonDefault[stats](tx, statsKey, stats{})
			if err != nil {
				return fmt.Errorf("get nc stats: %w", err)
			}
			if stats.Solutions == nil {
				stats.Solutions = make(map[solutionKey]solution)
			}
			if stats.DaysInfo == nil {
				stats.DaysInfo = make(map[int64]statsDayInfo)
			}

			msgToDayIdx, err := db.GetJsonDefault(tx, msgToDayInfoKey, make(map[int]statsDayInfo))
			if err != nil {
				return fmt.Errorf("get msgToDayIdx: %w", err)
			}

			dayInfo, ok := msgToDayIdx[msg.ReplyTo.ID]
			if !ok {
				return nil
			}

			key := solutionKey{
				DayIdx: dayInfo.DayIdx,
				UserID: sender.ID,
			}
			if _, ok := stats.Solutions[key]; ok {
				return set(tg.ReactionOk) // keep only first solution to not ruin solve time stats
			}

			stats.Solutions[key] = solution{Update: update}
			stats.DaysInfo[dayInfo.DayIdx] = dayInfo
			if err := db.SetJson(tx, statsKey, stats); err != nil {
				return fmt.Errorf("set stats: %w", err)
			}

			return set(tg.ReactionOk)
		})
	}
}

func (s *Service) publishRating(
	ctx context.Context,
	questionsToInclude int,
	header string,
	threadID int,
	msgToDayInfoKey string,
	statsKey string,
) error {
	return s.database.Do(ctx, func(tx db.Tx) error {
		lastDayInfo, err := s.getLastPublishedQuestionDayInfo(tx, msgToDayInfoKey)
		if err != nil {
			return fmt.Errorf("get last published question: %w", err)
		}

		stats, err := db.GetJson[stats](tx, statsKey)
		switch {
		case err == nil:
		case errors.Is(err, db.ErrKeyNotFound):
			return nil
		default:
			return fmt.Errorf("get stats: %w", err)
		}

		rating := buildRating(stats, lastDayInfo.DayIdx-int64(questionsToInclude)+1, lastDayInfo.DayIdx)
		if len(rating) == 0 {
			slog.Info("rating is empty skipping posting", slog.String("header", header))
			return nil
		}

		_, err = s.telegram.SendMarkdownV2(threadID, rating.toMarkdownV2(header))
		if err != nil {
			return fmt.Errorf("send rating: %w", err)
		}

		return nil
	})
}

type ratingRow struct {
	UserID        int64
	Username      string
	Name          string
	Solved        int
	CurrentStreak int
	MaxStreak     int
	SolveTime     time.Duration
}

func (r ratingRow) less(other ratingRow) bool {
	if r.Solved != other.Solved {
		return -r.Solved < -other.Solved
	}
	if r.CurrentStreak != other.CurrentStreak {
		return -r.CurrentStreak < -other.CurrentStreak
	}
	if r.MaxStreak != other.MaxStreak {
		return -r.MaxStreak < -other.MaxStreak
	}
	return r.SolveTime < other.SolveTime
}

type rating []ratingRow

func buildRating(stats stats, dayIdxFrom, dayIdxTo int64) rating {
	type userSolution struct {
		solutionKey
		solution
	}

	userSolutions := make(map[int64][]userSolution)
	for key, solution := range stats.Solutions {
		if !(key.DayIdx >= dayIdxFrom && key.DayIdx <= dayIdxTo) {
			continue
		}
		userSolutions[key.UserID] = append(userSolutions[key.UserID], userSolution{
			solutionKey: key,
			solution:    solution,
		})
	}

	result := make(rating, 0, len(userSolutions))
	for _, solutions := range userSolutions {
		sort.Slice(solutions, func(i, j int) bool {
			return solutions[i].DayIdx < solutions[j].DayIdx
		})

		row := ratingRow{}
		currIdx, currStreak, maxStreak := int64(0), 0, 0
		for _, sol := range solutions {
			msg := sol.Update.Message
			if msg == nil || msg.Sender == nil || msg.ReplyTo == nil {
				continue
			}
			row.UserID = sol.UserID
			row.Username = msg.Sender.Username
			row.Name = iter.JoinNonEmpty(" ", msg.Sender.FirstName, msg.Sender.LastName)
			row.Solved++
			row.SolveTime += msg.Time().Sub(msg.ReplyTo.Time())

			if sol.DayIdx-currIdx > 1 {
				maxStreak = max(maxStreak, currStreak)
				currStreak = 0
			}
			currStreak++
			currIdx = sol.DayIdx
		}
		if dayIdxTo-currIdx > 1 { // we allow gap of one because there's time after rating post and before next daily
			maxStreak = max(maxStreak, currStreak)
			currStreak = 0
		}
		row.CurrentStreak = currStreak
		row.MaxStreak = max(maxStreak, currStreak)
		result = append(result, row)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].less(result[j])
	})
	return result
}

func (r rating) toMarkdownV2(header string) string {
	var buf strings.Builder
	buf.Grow(len(header) + len(r)*30)
	buf.WriteString(tg.EscapeMD(header) + "\n")
	for _, row := range r {
		var name string
		if len(row.Username) > 0 {
			name = "@" + tg.EscapeMD(row.Username)
		} else {
			name = fmt.Sprintf("[%s](tg://user?id=%d)", tg.EscapeMD(row.Name), row.UserID)
		}
		solveTime := tg.EscapeMD(fmt.Sprintf("%.1fh", row.SolveTime.Hours()))
		buf.WriteString(fmt.Sprintf(
			"%s \\- solved %d, streak %d, max streak %d, total time %s\n",
			name, row.Solved, row.CurrentStreak, row.MaxStreak, solveTime,
		))
	}
	return buf.String()
}
