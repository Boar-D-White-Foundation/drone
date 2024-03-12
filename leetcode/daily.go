package leetcode

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os/exec"
	"time"

	"github.com/frosthamster/drone/retry"
)

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
		` --data-raw '{"query":"query questionOfToday {activeDailyCodingChallengeQuestion {date link}}","variables":{},"operationName":"questionOfToday"}'`
)

type dailyQuestionResp struct {
	Data struct {
		ActiveDailyCodingChallengeQuestion struct {
			Date string `json:"date"`
			Link string `json:"link"`
		} `json:"activeDailyCodingChallengeQuestion"`
	} `json:"data"`
}

func GetDailyLink(ctx context.Context) (string, error) {
	backoff := retry.LinearBackoff{
		Delay:       time.Second * 5,
		MaxAttempts: 10,
	}
	link, err := retry.Do(ctx, "lc daily fetch", backoff, func() (string, error) {
		var outBuf, errBuf bytes.Buffer
		cmd := exec.Command("bash", "-c", dailyQuestionCurlQuery)
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
		err := cmd.Run()
		if err != nil {
			slog.Error("failed to run curl", slog.String("stderr", errBuf.String()))
			return "", err
		}

		resp := dailyQuestionResp{}
		if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
			return "", err
		}

		link := resp.Data.ActiveDailyCodingChallengeQuestion.Link
		if len(link) == 0 {
			return "", ErrEmptyLink
		}
		return link, nil
	})
	if err != nil {
		return "", err
	}
	slog.Info("fetched lc daily", slog.String("link", link))

	return leetCodeUrl + link, nil
}
