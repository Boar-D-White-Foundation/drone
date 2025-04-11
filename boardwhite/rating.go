package boardwhite

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/leetcode"
	"github.com/boar-d-white-foundation/drone/retry"
	"github.com/boar-d-white-foundation/drone/tg"
	"golang.org/x/exp/slog"
	tele "gopkg.in/telebot.v3"
)

var estimatedComplexityRe = regexp.MustCompile(`[OoоО]\s?\((.+)\)[^OoоО]*[OoоО]\s?\((.+)\)`)

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

func okReaction(hasComplexityEstimate bool) tg.Reaction {
	if hasComplexityEstimate {
		return tg.ReactionFire
	}
	return tg.ReactionOk
}

func (s *Service) makeStatsHandler(
	pinnedMessagesKey string,
	msgToDayInfoKey string,
	statsKey string,
	ratingOpts ratingOpts,
) func(context.Context, tele.Context) error {
	return func(ctx context.Context, c tele.Context) error {
		update, msg, sender := c.Update(), c.Message(), c.Sender()
		if msg == nil || sender == nil {
			return nil
		}
		if msg.ReplyTo == nil || msg.ReplyTo.Sender.ID != c.Bot().Me.ID {
			return nil
		}

		set := tg.SetReactionFor(s.telegram, msg.ID)
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

				submissionID := leetcode.SubmissionID(match[1])
				backoff := retry.LinearBackoff{
					Delay:       time.Millisecond * 50,
					MaxAttempts: 3,
				}
				submission, err := retry.Do(
					ctx, "get lc submission "+submissionID.String(), backoff,
					func() (leetcode.Submission, error) {
						submission, err := s.lcClient.GetSubmission(ctx, submissionID)
						if err != nil {
							return leetcode.Submission{}, fmt.Errorf("get lc submission %s: %w", submissionID, err)
						}

						return submission, err
					},
				)
				if err != nil {
					return err
				}

				if !submission.IsSolved() {
					return set(tg.ReactionClown)
				}

				if s.mediaGenerator != nil {
					err := s.tasks.postCodeSnippet.Schedule(tx, 1, postCodeSnippetArgs{
						MessageID:  msg.ID,
						ThreadID:   msg.ThreadID,
						Submission: submission,
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

			oldSol, ok := stats.Solutions[key]
			hasComplexityEstimate := !ratingOpts.noComplexityEstimations && extractEstimatedComplexity(*msg).isFull()
			oldSolHasComplexityEstimate := !ratingOpts.noComplexityEstimations && oldSol.Update.Message != nil &&
				extractEstimatedComplexity(*oldSol.Update.Message).isFull()

			if ok && (ratingOpts.noComplexityEstimations || oldSolHasComplexityEstimate || !hasComplexityEstimate) {
				return set(okReaction(hasComplexityEstimate)) // keep only first solution to not ruin solve time stats
			}

			stats.Solutions[key] = solution{Update: update}
			stats.DaysInfo[dayInfo.DayIdx] = dayInfo
			if err := db.SetJson(tx, statsKey, stats); err != nil {
				return fmt.Errorf("set stats: %w", err)
			}

			return set(okReaction(hasComplexityEstimate))
		})
	}
}

func (s *Service) publishRating(
	ctx context.Context,
	questionsToInclude int,
	opts ratingOpts,
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

		rating := buildRating(stats, lastDayInfo.DayIdx-int64(questionsToInclude)+1, lastDayInfo.DayIdx, opts)
		if len(rating.rows) == 0 {
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
	Mention             string
	Solved              int
	CurrentStreak       int
	MaxStreak           int
	ComplexityEstimates int
	SolveTime           time.Duration
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
	if r.ComplexityEstimates != other.ComplexityEstimates {
		return -r.ComplexityEstimates < -other.ComplexityEstimates
	}
	return r.SolveTime < other.SolveTime
}

type ratingOpts struct {
	noComplexityEstimations bool
}

type rating struct {
	rows []ratingRow
	opts ratingOpts
}

type complexity struct {
	time   string
	memory string
}

func (c complexity) isFull() bool {
	return len(c.time) > 0 && len(c.memory) > 0
}

func extractEstimatedComplexity(msg tele.Message) complexity {
	var text string
	if len(msg.Caption) > 0 {
		text = msg.Caption
	} else {
		text = msg.Text
	}
	match := estimatedComplexityRe.FindStringSubmatch(text)
	if len(match) < 3 {
		return complexity{}
	}

	return complexity{
		time:   match[1],
		memory: match[2],
	}
}

func buildRating(stats stats, dayIdxFrom, dayIdxTo int64, opts ratingOpts) rating {
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

	rows := make([]ratingRow, 0, len(userSolutions))
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
			row.Mention = tg.BuildMentionMarkdownV2(*msg.Sender)
			row.Solved++
			row.SolveTime += msg.Time().Sub(msg.ReplyTo.Time())
			if !opts.noComplexityEstimations && extractEstimatedComplexity(*msg).isFull() {
				row.ComplexityEstimates++
			}

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
		rows = append(rows, row)
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].less(rows[j])
	})
	return rating{
		rows: rows,
		opts: opts,
	}
}

func (r rating) toMarkdownV2(header string) string {
	var buf strings.Builder
	buf.Grow(len(header) + len(r.rows)*30)
	buf.WriteString(tg.EscapeMD(header) + "\n")
	for _, row := range r.rows {
		solveTime := tg.EscapeMD(fmt.Sprintf("%.1fh", row.SolveTime.Hours()))
		buf.WriteString(fmt.Sprintf(
			"%s \\- solved %d, streak %d, max streak %d, ",
			row.Mention, row.Solved, row.CurrentStreak, row.MaxStreak,
		))
		if !r.opts.noComplexityEstimations {
			buf.WriteString(tg.EscapeMD(fmt.Sprintf("O(f) estimates %d, ", row.ComplexityEstimates)))
		}
		buf.WriteString(fmt.Sprintf("total time %s\n", solveTime))
	}
	return buf.String()
}
