package boardwhite

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/boar-d-white-foundation/drone/alert"
	"github.com/boar-d-white-foundation/drone/config"
	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/dbq"
	"github.com/boar-d-white-foundation/drone/leetcode"
	"github.com/boar-d-white-foundation/drone/media"
	"github.com/boar-d-white-foundation/drone/tg"
)

const (
	keyLCPinnedMessages       = "boardwhite:leetcode:pinned_messages"
	keyLCPinnedToStatsDayInfo = "boardwhite:leetcode:pinned_to_stats_day_info"
	keyLCStats                = "boardwhite:leetcode:stats"

	keyLCChickensPinnedMessages       = "boardwhite:leetcode_chickens:pinned_messages"
	keyLCChickensPinnedToStatsDayInfo = "boardwhite:leetcode_chickens:pinned_to_stats_day_info"
	keyLCChickensStats                = "boardwhite:leetcode_chickens:stats"
	keyLCChickensFallbackQuestionIdx  = "boardwhite:leetcode_chickens:fallback_question_idx"

	keyNCPinnedMessages       = "boardwhite:neetcode:pinned_messages"
	keyNCPinnedToStatsDayInfo = "boardwhite:neetcode:pinned_to_stats_day_info"
	keyNCStats                = "boardwhite:neetcode:stats"

	keyOnJoinGreetedUsers = "boardwhite:on_join_greeted_users"

	keyOboronaLastGeneratedAt = "boardwhite:oborona:last_generated_at"

	keyOkrValues        = "boardwhite:okr:values"
	keyOkrPinnedMessage = "boardwhite:okr:pinned_message"
)

var (
	lcSubmissionRe = regexp.MustCompile(`https://leetcode\.com.*/submissions/(?:detail/)?(?P<submissionID>\d+)`)
)

type MockConfig struct {
	Period     time.Duration
	StickerIDs []string
}

type Config struct {
	ChatID                     int64
	LeetcodeThreadID           int
	LeetcodeChickensThreadID   int
	DailyStickersIDs           []string
	DailyChickensStickerIDs    []string
	DpStickerID                string
	Mocks                      map[string]MockConfig
	FloodThreadID              int
	GreetingsNewUsersTemplates []string
	GreetingsOldUsersTemplates []string
	OboronaPeriod              time.Duration
	OboronaTemplate            string
	OboronaWords               [][]string
	InterviewsThreadID         int
}

type tasks struct {
	postCodeSnippet dbq.Task[postCodeSnippetArgs]
}

type Service struct {
	cfg                Config
	tasks              tasks
	database           db.DB
	telegram           tg.Client
	lcChickenQuestions lcChickenQuestions
	alerts             *alert.Manager
	mediaGenerator     *media.Generator // can be nil
	lcClient           *leetcode.Client
	vcLinkRe           *regexp.Regexp
}

func NewService(
	cfg Config,
	telegram tg.Client,
	database db.DB,
	alerts *alert.Manager,
	mediaGenerator *media.Generator,
	lcClient *leetcode.Client,
	vcLinkRe *regexp.Regexp,
) (*Service, error) {
	questions, err := newLCChickenQuestions()
	if err != nil {
		return nil, fmt.Errorf("load lcChickenQuestions: %w", err)
	}

	return &Service{
		cfg:                cfg,
		database:           database,
		telegram:           telegram,
		lcChickenQuestions: questions,
		alerts:             alerts,
		mediaGenerator:     mediaGenerator,
		lcClient:           lcClient,
		vcLinkRe:           vcLinkRe,
	}, nil
}

func NewServiceFromConfig(
	cfg config.Config,
	telegram tg.Client,
	database db.DB,
	alerts *alert.Manager,
	mediaGenerator *media.Generator,
	lcClient *leetcode.Client,
) (*Service, error) {
	mocks := make(map[string]MockConfig)
	for _, v := range cfg.Mocks {
		period, err := time.ParseDuration(v.Period)
		if err != nil {
			return nil, fmt.Errorf("parse duration %q: %w", v.Period, err)
		}
		mocks[v.Username] = MockConfig{
			Period:     period,
			StickerIDs: v.StickerIDs,
		}
	}

	vcLinkRe := regexp.MustCompile(fmt.Sprintf(
		`(?P<link>https://%s/[^/]+/\d+)`,
		strings.Replace(cfg.VC.Domain, ".", `\.`, -1),
	))
	serviceCfg := Config{
		ChatID:                     cfg.Boardwhite.ChatID,
		LeetcodeThreadID:           cfg.Boardwhite.LeetCodeThreadID,
		LeetcodeChickensThreadID:   cfg.Boardwhite.LeetcodeChickensThreadID,
		DailyStickersIDs:           cfg.DailyStickerIDs,
		DailyChickensStickerIDs:    cfg.DailyChickensStickerIDs,
		DpStickerID:                cfg.DPStickerID,
		Mocks:                      mocks,
		FloodThreadID:              cfg.Boardwhite.FloodThreadID,
		GreetingsNewUsersTemplates: cfg.GreetingsNewUsersTemplates,
		GreetingsOldUsersTemplates: cfg.GreetingsOldUsersTemplates,
		OboronaPeriod:              cfg.Oborona.Period,
		OboronaTemplate:            cfg.Oborona.Template,
		OboronaWords:               cfg.Oborona.Words,
		InterviewsThreadID:         cfg.Boardwhite.InterviewsThreadID,
	}
	return NewService(serviceCfg, telegram, database, alerts, mediaGenerator, lcClient, vcLinkRe)
}

type publishDailyReq struct {
	dayIdx          int64
	threadID        int
	header          string
	text            string
	stickerID       string
	pinnedMsgsKey   string
	msgToDayInfoKey string
}

func (s *Service) publishDaily(tx db.Tx, req publishDailyReq) (int, error) {
	pinnedIDs, err := db.GetJsonDefault[[]int](tx, req.pinnedMsgsKey, nil)
	if err != nil {
		return 0, fmt.Errorf("get key %s: %w", req.pinnedMsgsKey, err)
	}
	if len(pinnedIDs) > 0 {
		// last is considered active
		err = s.telegram.Unpin(pinnedIDs[len(pinnedIDs)-1])
		if err != nil {
			slog.Error("err unpin", slog.Any("err", err))
		}
	}

	messageID, err := s.telegram.SendSpoilerLink(req.threadID, req.header, req.text)
	if err != nil {
		return 0, fmt.Errorf("send daily: %w", err)
	}

	_, err = s.telegram.SendSticker(req.threadID, req.stickerID)
	if err != nil {
		return 0, fmt.Errorf("send sticker: %w", err)
	}

	err = s.telegram.Pin(messageID)
	if err != nil {
		return 0, fmt.Errorf("pin: %w", err)
	}

	pinnedIDs = append(pinnedIDs, messageID)
	if err := db.SetJson(tx, req.pinnedMsgsKey, pinnedIDs); err != nil {
		return 0, fmt.Errorf("set key %s: %w", req.pinnedMsgsKey, err)
	}

	msgToDayInfo, err := db.GetJsonDefault(tx, req.msgToDayInfoKey, make(map[int]statsDayInfo))
	if err != nil {
		return 0, fmt.Errorf("get msgToDayInfo: %w", err)
	}

	msgToDayInfo[messageID] = statsDayInfo{
		DayIdx:      req.dayIdx,
		PublishedAt: time.Now(),
	}
	if err := db.SetJson(tx, req.msgToDayInfoKey, msgToDayInfo); err != nil {
		return 0, fmt.Errorf("set msgToDayInfo: %w", err)
	}

	return messageID, nil
}
