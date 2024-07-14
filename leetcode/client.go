package leetcode

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/boar-d-white-foundation/drone/config"
)

const (
	querySubmission = `
		query submissionDetails($submissionId: Int!) {
			submissionDetails(submissionId: $submissionId) {
				runtime
				runtimePercentile
				memory
				memoryPercentile
				code
				statusCode
				lang {
					name
				}
			}
		}
	`
)

type Config struct {
	Session string
	CSRF    string
}

type Client struct {
	cfg    Config
	client http.Client
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg:    cfg,
		client: http.Client{Timeout: 5 * time.Second},
	}
}

func NewClientFromConfig(cfg config.Config) *Client {
	return NewClient(Config{
		Session: cfg.Tg.Session,
		CSRF:    cfg.Tg.CSRF,
	})
}

type gqReq struct {
	Method string `json:"submissionDetails"`
	Query  string `json:"query"`
	Args   any    `json:"variables,omitempty"`
}

type submission struct {
	Data struct {
		SubmissionDetails struct {
			Runtime           int     `json:"runtime"`
			RuntimePercentile float64 `json:"runtimePercentile"`
			Memory            int     `json:"memory"`
			MemoryPercentile  float64 `json:"memoryPercentile"`
			Code              string  `json:"code"`
			StatusCode        int     `json:"statusCode"`
			Lang              struct {
				Name string `json:"name"`
			} `json:"lang"`
		} `json:"submissionDetails"`
	} `json:"data"`
}

type Submission struct {
	Runtime           int
	RuntimePercentile float64
	Memory            int
	MemoryPercentile  float64
	Code              string
	Lang              string
}

func (c *Client) GetSubmission(ctx context.Context, id string) (Submission, error) {
	type args struct {
		SubmissionID string `json:"submissionId"`
	}
	gqr := gqReq{
		Method: "submissionDetails",
		Query:  querySubmission,
		Args:   args{SubmissionID: id},
	}
	body, err := json.Marshal(gqr)
	if err != nil {
		return Submission{}, fmt.Errorf("marshall submissionDetails body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://leetcode.com/graphql", bytes.NewReader(body))
	if err != nil {
		return Submission{}, fmt.Errorf("create request: %w", err)
	}

	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Csrftoken", c.cfg.CSRF)
	req.Header.Add("Referer", "https://leetcode.com")
	req.AddCookie(&http.Cookie{Name: "LEETCODE_SESSION", Value: c.cfg.Session})
	req.AddCookie(&http.Cookie{Name: "csrftoken", Value: c.cfg.CSRF})
	resp, err := c.client.Do(req)
	if err != nil {
		return Submission{}, fmt.Errorf("send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("close submissionDetails resp body", slog.Any("err", err))
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Submission{}, fmt.Errorf("read submissionDetails resp body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return Submission{}, fmt.Errorf("got non %d resp for submissionDetails: %s", resp.StatusCode, string(respBody))
	}

	raw := submission{}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return Submission{}, fmt.Errorf("unmarshal submissionDetails resp body: %w", err)
	}
	if len(raw.Data.SubmissionDetails.Code) == 0 {
		return Submission{}, errors.New("got empty code, probably session or csrf are expired or empty")
	}

	return Submission{
		Runtime:           raw.Data.SubmissionDetails.Runtime,
		RuntimePercentile: raw.Data.SubmissionDetails.RuntimePercentile,
		Memory:            raw.Data.SubmissionDetails.Memory,
		MemoryPercentile:  raw.Data.SubmissionDetails.MemoryPercentile,
		Code:              raw.Data.SubmissionDetails.Code,
		Lang:              raw.Data.SubmissionDetails.Lang.Name,
	}, nil
}
