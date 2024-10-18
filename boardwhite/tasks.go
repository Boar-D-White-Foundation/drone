package boardwhite

import (
	"bytes"
	"context"
	"fmt"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/dbq"
	"github.com/boar-d-white-foundation/drone/leetcode"
)

func (s *Service) RegisterTasks(registry *dbq.Registry) error {
	postCodeSnippetTask, err := dbq.RegisterHandler(registry, "boardwhite:post_code_snippet", s.postCodeSnippet)
	if err != nil {
		return fmt.Errorf("register post code snippet taskl: %w", err)
	}

	s.tasks.postCodeSnippet = postCodeSnippetTask
	return nil
}

type postCodeSnippetArgs struct {
	MessageID  int                 `json:"message_id"`
	ThreadID   int                 `json:"thread_id"`
	Submission leetcode.Submission `json:"submission"`
}

func (s *Service) postCodeSnippet(ctx context.Context, tx db.Tx, args postCodeSnippetArgs) error {
	sub := args.Submission
	snippet, err := s.imageGenerator.GenerateCodeSnippet(ctx, sub.ID, sub.Lang, sub.Code)
	if err != nil {
		s.alerts.Errorxf(err, "err generate snippet: %+v", args)
		return fmt.Errorf("generate snippet: %w", err)
	}

	caption := ""
	if args.ThreadID != s.cfg.LeetcodeChickensThreadID {
		caption = fmt.Sprintf(
			"Runtime beats %.0f%%\nMemory beats %.0f%%",
			sub.RuntimePercentile, sub.MemoryPercentile,
		)
	}
	imgName := fmt.Sprintf("submission_%s.png", sub.ID)
	_, err = s.telegram.ReplyWithSpoilerPhoto(
		args.MessageID,
		caption,
		imgName,
		"image/png",
		bytes.NewReader(snippet),
	)
	if err != nil {
		s.alerts.Errorxf(err, "err reply with snippet: %+v", args)
		return fmt.Errorf("reply with snippet: %w", err)
	}

	return nil
}
