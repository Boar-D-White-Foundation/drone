package boardwhite

import (
	"bytes"
	"context"
	"fmt"

	"github.com/boar-d-white-foundation/drone/db"
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

func (s *Service) postCodeSnippet(ctx context.Context, tx db.Tx, args postCodeSnippetArgs) error {
	submission, err := s.lcClient.GetSubmission(ctx, args.SubmissionID)
	if err != nil {
		s.alerts.Errorxf(err, "err get submission: %+v", args)
		return fmt.Errorf("get submission: %w", err)
	}

	snippet, err := s.imageGenerator.GenerateCodeSnippet(ctx, args.SubmissionID, submission.Lang, submission.Code)
	if err != nil {
		s.alerts.Errorxf(err, "err generate snippet: %+v", args)
		return fmt.Errorf("generate snippet: %w", err)
	}

	_, err = s.telegram.ReplyWithSpoilerPhoto(
		args.MessageID,
		fmt.Sprintf("submission_%s.png", args.SubmissionID),
		"image/png",
		bytes.NewReader(snippet),
	)
	if err != nil {
		s.alerts.Errorxf(err, "err reply with snippet: %+v", args)
		return fmt.Errorf("reply with snippet: %w", err)
	}

	return nil
}
