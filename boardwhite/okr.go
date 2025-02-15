package boardwhite

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"text/template"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/tg"
	tele "gopkg.in/telebot.v3"
)

type okrTag string

func (t okrTag) String() string {
	return string(t)
}

var (
	okrTagUnfortunately    okrTag = "#unfortunately2025"
	okrTagTier1000Offer    okrTag = "#tier1000_offer2025"
	okrTagBigtechOffer     okrTag = "#bigtech_offer2025"
	okrTagFaangOffer       okrTag = "#faang_offer2025"
	okrTagSeniorPromo      okrTag = "#senior_promo2025"
	okrTagStaffPromo       okrTag = "#staff_promo2025"
	okrTagRussiaRelocation okrTag = "#russia2025"
	okrTagUsaRelocation    okrTag = "#usa2025"
)

const okrRemoveCommand = "/remove_okr"

type okrGoal struct {
	Tag  okrTag
	Goal int
}

var okrGoals = map[okrTag]okrGoal{
	okrTagUnfortunately: {
		Tag:  okrTagUnfortunately,
		Goal: 300,
	},
	okrTagTier1000Offer: {
		Tag:  okrTagTier1000Offer,
		Goal: 50,
	},
	okrTagBigtechOffer: {
		Tag:  okrTagBigtechOffer,
		Goal: 5,
	},
	okrTagFaangOffer: {
		Tag:  okrTagFaangOffer,
		Goal: 3,
	},
	okrTagSeniorPromo: {
		Tag:  okrTagSeniorPromo,
		Goal: 3,
	},
	okrTagStaffPromo: {
		Tag:  okrTagStaffPromo,
		Goal: 1,
	},
	okrTagRussiaRelocation: {
		Tag:  okrTagRussiaRelocation,
		Goal: 1,
	},
	okrTagUsaRelocation: {
		Tag:  okrTagUsaRelocation,
		Goal: 1,
	},
}

type okrProgress struct {
	Current int
	Goal    int
	Tag     okrTag
	Status  string
}

type okrTemplateData struct {
	Bigtech       okrProgress
	Faang         okrProgress
	Tier1000      okrProgress
	Senior        okrProgress
	Staff         okrProgress
	Usa           okrProgress
	Russia        okrProgress
	Unfortunately okrProgress
}

type okrs struct {
	// we need it to preserve old data without saved updates,
	// but it should be close to sum(sum(u.Counts.values()) for u in Updates)
	TotalCount map[okrTag]int `json:"total_count"`
	Updates    []okrUpdate    `json:"updates"`
}

func initOkrs(okrs *okrs) {
	if okrs.TotalCount == nil {
		okrs.TotalCount = make(map[okrTag]int)
	}
	for i := range okrs.Updates {
		if okrs.Updates[i].Counts == nil {
			okrs.Updates[i].Counts = make(map[okrTag]int)
		}
	}
}

// one update might contain same tag multiple times,
// so we need to keep track what count of the particular okr tag
// is bound to the particular update
// Counts holds actual count with respect to removes
type okrUpdate struct {
	Update tele.Update    `json:"update"`
	Counts map[okrTag]int `json:"counts"`
}

var okrMessageTemplate = template.Must(template.New("okr").Parse(`ОКРы 2025:

Офферы:
{{.Tier1000.Current}}/{{.Tier1000.Goal}} в тир 1000 ({{.Tier1000.Tag}}) {{.Tier1000.Status}}
{{.Bigtech.Current}}/{{.Bigtech.Goal}} в бигтех ({{.Bigtech.Tag}}) {{.Bigtech.Status}}
{{.Faang.Current}}/{{.Faang.Goal}} в FAANG ({{.Faang.Tag}}) {{.Faang.Status}}

Промо:
{{.Senior.Current}}/{{.Senior.Goal}} на сеньора ({{.Senior.Tag}}) {{.Senior.Status}}
{{.Staff.Current}}/{{.Staff.Goal}} на стаффа ({{.Staff.Tag}}) {{.Staff.Status}}

Релокация:
{{.Russia.Current}}/{{.Russia.Goal}} релока в Россию ({{.Russia.Tag}}) {{.Russia.Status}}
{{.Usa.Current}}/{{.Usa.Goal}} релока в США ({{.Usa.Tag}}) {{.Usa.Status}}

Анфочантли:
{{.Unfortunately.Current}}/{{.Unfortunately.Goal}} ({{.Unfortunately.Tag}}) {{.Unfortunately.Status}}`))

func (s *Service) OnUpdateOkr(ctx context.Context, c tele.Context) error {
	msg, chat, update := c.Message(), c.Chat(), c.Update()
	if msg == nil || chat == nil || chat.ID != s.cfg.ChatID || msg.IsForwarded() {
		return nil
	}
	if strings.HasPrefix(msg.Text, okrRemoveCommand) {
		return nil
	}

	counts := extractOkrTagsCounts(msg.Text)
	if len(counts) == 0 {
		return nil
	}

	set := tg.SetReactionFor(s.telegram, msg.ID)
	return s.database.Do(ctx, func(tx db.Tx) error {
		okrs, err := db.GetJsonDefault(tx, keyOkrValues, okrs{})
		if err != nil {
			return fmt.Errorf("get okr values: %w", err)
		}
		initOkrs(&okrs)

		for tag, count := range counts {
			okrs.TotalCount[tag] += count
		}
		okrs.Updates = append(okrs.Updates, okrUpdate{
			Update: update,
			Counts: counts,
		})

		if err := s.saveOkrsAndUpsertTgMsg(tx, okrs); err != nil {
			return fmt.Errorf("save okrs and upsert tg msg: %w", err)
		}

		return set(tg.ReactionWriting)
	})
}

func (s *Service) OnRemoveOkr(ctx context.Context, c tele.Context) error {
	msg, chat := c.Message(), c.Chat()
	if msg == nil || chat == nil || chat.ID != s.cfg.ChatID {
		return nil
	}
	if !strings.HasPrefix(msg.Text, okrRemoveCommand) {
		return nil
	}

	set := tg.SetReactionFor(s.telegram, msg.ID)
	if msg.ReplyTo == nil {
		return set(tg.ReactionClown)
	}

	countsToRemove := extractOkrTagsCounts(msg.Text)
	removeAll := msg.Text == okrRemoveCommand
	return s.database.Do(ctx, func(tx db.Tx) error {
		okrs, err := db.GetJsonDefault(tx, keyOkrValues, okrs{})
		if err != nil {
			return fmt.Errorf("get okr values: %w", err)
		}
		initOkrs(&okrs)

		idx := slices.IndexFunc(okrs.Updates, func(update okrUpdate) bool {
			return update.Update.Message.ID == msg.ReplyTo.ID
		})
		if idx == -1 {
			return set(tg.ReactionClown)
		}

		update := okrs.Updates[idx]
		if removeAll {
			countsToRemove = update.Counts
		}

		for tag, removeCount := range countsToRemove {
			if update.Counts[tag] < removeCount {
				return set(tg.ReactionClown)
			}

			okrs.TotalCount[tag] -= removeCount
			update.Counts[tag] -= removeCount
			if update.Counts[tag] == 0 {
				delete(update.Counts, tag)
			}
		}

		okrs.Updates[idx] = update
		if len(update.Counts) == 0 {
			okrs.Updates = slices.Delete(okrs.Updates, idx, idx+1)
		}

		if err := s.saveOkrsAndUpsertTgMsg(tx, okrs); err != nil {
			return fmt.Errorf("save okrs and upsert tg msg: %w", err)
		}

		return set(tg.ReactionWriting)
	})
}

func (s *Service) saveOkrsAndUpsertTgMsg(tx db.Tx, okrs okrs) error {
	if err := db.SetJson(tx, keyOkrValues, okrs); err != nil {
		return fmt.Errorf("save okrs: %w", err)
	}

	progressMsg, err := buildOkrProgressMsg(okrs.TotalCount)
	if err != nil {
		return fmt.Errorf("construct progress message: %w", err)
	}

	if err := s.upsertPinnedOkrMsg(tx, progressMsg); err != nil {
		return fmt.Errorf("upsert pinned okr message: %w", err)
	}

	return nil
}

func extractOkrTagsCounts(msgText string) map[okrTag]int {
	tags := make(map[okrTag]int)
	for tag := range okrGoals {
		if occurrences := strings.Count(msgText, tag.String()); occurrences > 0 {
			tags[tag] = occurrences
		}
	}

	return tags
}

func (s *Service) upsertPinnedOkrMsg(tx db.Tx, progressMessage string) error {
	pinnedMsgID, err := db.GetJson[int](tx, keyOkrPinnedMessage)
	if errors.Is(err, db.ErrKeyNotFound) {
		if err := s.postNewOkrMessage(tx, progressMessage); err != nil {
			return fmt.Errorf("post initial okr message: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("get pinned message id: %w", err)
	}

	if err := s.telegram.EditMessageText(pinnedMsgID, progressMessage); err != nil {
		return fmt.Errorf("edit pinned okr message: %w", err)
	}

	return nil
}

func (s *Service) postNewOkrMessage(tx db.Tx, progressMessage string) error {
	messageID, err := s.telegram.SendText(s.cfg.InterviewsThreadID, progressMessage)
	if err != nil {
		return fmt.Errorf("send okr message: %w", err)
	}

	if err := s.telegram.Pin(messageID); err != nil {
		return fmt.Errorf("pin okr message: %w", err)
	}

	if err := db.SetJson(tx, keyOkrPinnedMessage, messageID); err != nil {
		return fmt.Errorf("save new pinned okr message id: %w", err)
	}

	return nil
}

func buildOkrProgressMsg(counts map[okrTag]int) (string, error) {
	build := func(tag okrTag) okrProgress {
		goal := okrGoals[tag]
		count := counts[tag]
		return okrProgress{
			Current: count,
			Goal:    goal.Goal,
			Tag:     tag,
			Status:  getStatusEmoji(count, goal.Goal),
		}
	}

	data := okrTemplateData{
		Bigtech:       build(okrTagBigtechOffer),
		Faang:         build(okrTagFaangOffer),
		Tier1000:      build(okrTagTier1000Offer),
		Senior:        build(okrTagSeniorPromo),
		Staff:         build(okrTagStaffPromo),
		Usa:           build(okrTagUsaRelocation),
		Russia:        build(okrTagRussiaRelocation),
		Unfortunately: build(okrTagUnfortunately),
	}

	var buf bytes.Buffer
	if err := okrMessageTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute okr template: %w", err)
	}

	return buf.String(), nil
}

func getStatusEmoji(count, goal int) string {
	if count >= goal {
		return "✅"
	}
	return "⏳"
}
