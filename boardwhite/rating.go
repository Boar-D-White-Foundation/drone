package boardwhite

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/boar-d-white-foundation/drone/iter"
	"github.com/boar-d-white-foundation/drone/tg"
	tele "gopkg.in/telebot.v3"
)

type solutionKey struct {
	DayIdx int64 `json:"day_idx"`
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

type stats struct {
	Solutions map[solutionKey]solution `json:"solutions,omitempty"`
}

func newStats() stats {
	return stats{
		Solutions: make(map[solutionKey]solution),
	}
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

func buildRating(stats stats) rating {
	type userSolution struct {
		solutionKey
		solution
	}

	userSolutions := make(map[int64][]userSolution)
	for key, solution := range stats.Solutions {
		userSolutions[key.UserID] = append(userSolutions[key.UserID], userSolution{
			solutionKey: key,
			solution:    solution,
		})
	}

	result := make(rating, 0, len(userSolutions))
	currentDayIdx := int64(0)
	for _, solutions := range userSolutions {
		for _, sol := range solutions {
			currentDayIdx = max(currentDayIdx, sol.DayIdx)
		}
	}
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
		if currentDayIdx-currIdx > 1 {
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
