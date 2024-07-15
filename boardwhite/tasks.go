package boardwhite

import (
	"bytes"
	"context"
	"fmt"

	"github.com/boar-d-white-foundation/drone/chrome"
	"github.com/boar-d-white-foundation/drone/dbq"
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
	MessageID    int    `json:"message_id"`
	SubmissionID string `json:"submission_id"`
}

func (s *Service) postCodeSnippet(ctx context.Context, args postCodeSnippetArgs) error {
	sub, err := s.lcClient.GetSubmission(ctx, args.SubmissionID)
	if err != nil {
		return fmt.Errorf("get submission: %w", err)
	}

	snippet, err := chrome.GenerateCodeSnippet(ctx, s.browser, args.SubmissionID, sub.Code)
	if err != nil {
		return fmt.Errorf("generate snippet: %w", err)
	}

	_, err = s.telegram.ReplyWithSpoilerPhoto(
		args.MessageID,
		fmt.Sprintf("submission_%s.png", args.SubmissionID),
		"image/png",
		bytes.NewReader(snippet),
	)
	return err
}
