package leetcode

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os/exec"
	"time"

	"github.com/boar-d-white-foundation/drone/retry"
)

type Difficulty int

const (
	DifficultyUnknown Difficulty = iota
	DifficultyEasy
	DifficultyMedium
	DifficultyHard
)

func NewDifficulty(raw string) Difficulty {
	switch {
	case raw == "Easy":
		return DifficultyEasy
	case raw == "Medium":
		return DifficultyMedium
	case raw == "Hard":
		return DifficultyHard
	default:
		return DifficultyUnknown
	}
}

func (d Difficulty) String() string {
	switch d {
	case DifficultyEasy:
		return "Easy"
	case DifficultyMedium:
		return "Medium"
	case DifficultyHard:
		return "Hard"
	default:
		return ""
	}
}

func (d Difficulty) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *Difficulty) UnmarshalText(data []byte) error {
	*d = NewDifficulty(string(data))
	return nil
}

var (
	ErrEmptyLink = errors.New("got empty link")
)

const (
	leetCodeUrl = "https://leetcode.com"
	// leetcode probably uses tls fingerprinting to filter out non-standard http clients, so we can't use go http.Client
	// but for some reason curl works quite ok, so we just use it via os exec
	// as an alternative we can explore usage of uTLS to mimic as chrome
	dailyQuestionCurlQuery = `curl` +
		` -vvv --ipv4 -X POST 'https://leetcode.com/graphql/'` +
		` -H 'Content-type: application/json'` +
		` -H 'Origin: leetcode.com'` +
		` -H 'User-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36'` +
		` --data-raw '{"query":"query questionOfToday {activeDailyCodingChallengeQuestion {date link question {difficulty}}}","variables":{},"operationName":"questionOfToday"}'`
)

type dailyQuestionResp struct {
	Data struct {
		ActiveDailyCodingChallengeQuestion struct {
			Date     string `json:"date"`
			Link     string `json:"link"`
			Question struct {
				Difficulty string `json:"difficulty"`
			} `json:"question"`
		} `json:"activeDailyCodingChallengeQuestion"`
	} `json:"data"`
}

type DailyInfo struct {
	Link       string
	Difficulty Difficulty
}

func GetDailyInfo(ctx context.Context) (DailyInfo, error) {
	backoff := retry.LinearBackoff{
		Delay:       time.Second * 5,
		MaxAttempts: 10,
	}
	resp, err := retry.Do(ctx, "lc daily fetch", backoff, func() (dailyQuestionResp, error) {
		var outBuf, errBuf bytes.Buffer
		cmd := exec.Command("bash", "-c", dailyQuestionCurlQuery)
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
		if err := cmd.Run(); err != nil {
			slog.Error("failed to run curl", slog.String("stderr", errBuf.String()))
			return dailyQuestionResp{}, err
		}

		resp := dailyQuestionResp{}
		if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
			return dailyQuestionResp{}, err
		}

		link := resp.Data.ActiveDailyCodingChallengeQuestion.Link
		if len(link) == 0 {
			return dailyQuestionResp{}, ErrEmptyLink
		}
		return resp, nil
	})
	if err != nil {
		return DailyInfo{}, err
	}
	slog.Info("fetched lc daily", slog.Any("resp", resp))

	result := DailyInfo{
		Link:       leetCodeUrl + resp.Data.ActiveDailyCodingChallengeQuestion.Link,
		Difficulty: NewDifficulty(resp.Data.ActiveDailyCodingChallengeQuestion.Question.Difficulty),
	}
	return result, nil
}
