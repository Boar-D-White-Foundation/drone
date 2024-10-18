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
				totalCorrect
        		totalTestcases
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
	client *http.Client
}

func NewClient(cfg Config) *Client {
	client := http.Client{Timeout: 5 * time.Second}
	return &Client{
		cfg:    cfg,
		client: &client,
	}
}

func NewClientFromConfig(cfg config.Config) *Client {
	return NewClient(Config{
		Session: cfg.Leetcode.Session,
		CSRF:    cfg.Leetcode.CSRF,
	})
}

type Lang int

const (
	LangUnknown Lang = iota
	LangCPP
	LangJava
	LangPy2
	LangPy3
	LangC
	LangCSharp
	LangJS
	LangTS
	LangPHP
	LangSwift
	LangKotlin
	LangGO
	LangRuby
	LangScala
	LangRust
	LangRacket
)

func NewLang(raw string) Lang {
	switch {
	case raw == "cpp":
		return LangCPP
	case raw == "java":
		return LangJava
	case raw == "python":
		return LangPy2
	case raw == "python3":
		return LangPy3
	case raw == "c":
		return LangC
	case raw == "csharp":
		return LangCSharp
	case raw == "javascript":
		return LangJS
	case raw == "typescript":
		return LangTS
	case raw == "php":
		return LangPHP
	case raw == "swift":
		return LangSwift
	case raw == "kotlin":
		return LangKotlin
	case raw == "golang":
		return LangGO
	case raw == "ruby":
		return LangRuby
	case raw == "scala":
		return LangScala
	case raw == "rust":
		return LangRust
	case raw == "racket":
		return LangRacket
	default:
		return LangUnknown
	}
}

func (l *Lang) UnmarshalText(data []byte) error {
	*l = NewLang(string(data))
	return nil
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
			TotalCorrect      int     `json:"totalCorrect"`
			TotalTestcases    int     `json:"totalTestcases"`
			Lang              struct {
				Name Lang `json:"name"`
			} `json:"lang"`
		} `json:"submissionDetails"`
	} `json:"data"`
}

type SubmissionID string

func (sid SubmissionID) String() string {
	return string(sid)
}

type Submission struct {
	ID                SubmissionID `json:"id"`
	Runtime           int          `json:"runtime"`
	RuntimePercentile float64      `json:"runtime_percentile"`
	Memory            int          `json:"memory"`
	MemoryPercentile  float64      `json:"memory_percentile"`
	Code              string       `json:"code"`
	Lang              Lang         `json:"lang"`
	TotalCorrect      int          `json:"total_correct"`
	TotalTestcases    int          `json:"total_testcases"`
}

func (s Submission) IsSolved() bool {
	return s.TotalCorrect == s.TotalTestcases
}

func (c *Client) GetSubmission(ctx context.Context, id SubmissionID) (Submission, error) {
	type args struct {
		SubmissionID SubmissionID `json:"submissionId"`
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://leetcode.com/graphql", bytes.NewReader(body))
	if err != nil {
		return Submission{}, fmt.Errorf("create request: %w", err)
	}

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
		return Submission{}, fmt.Errorf(
			"got non 200 resp for submissionDetails %d: %s",
			resp.StatusCode, string(respBody),
		)
	}

	raw := submission{}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return Submission{}, fmt.Errorf("unmarshal submissionDetails resp body: %w", err)
	}
	if len(raw.Data.SubmissionDetails.Code) == 0 {
		return Submission{}, errors.New("got empty code, probably session or csrf are expired or empty")
	}

	return Submission{
		ID:                id,
		Runtime:           raw.Data.SubmissionDetails.Runtime,
		RuntimePercentile: raw.Data.SubmissionDetails.RuntimePercentile,
		Memory:            raw.Data.SubmissionDetails.Memory,
		MemoryPercentile:  raw.Data.SubmissionDetails.MemoryPercentile,
		Code:              raw.Data.SubmissionDetails.Code,
		Lang:              raw.Data.SubmissionDetails.Lang.Name,
		TotalCorrect:      raw.Data.SubmissionDetails.TotalCorrect,
		TotalTestcases:    raw.Data.SubmissionDetails.TotalTestcases,
	}, nil
}
