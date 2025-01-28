package boardwhite

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/boar-d-white-foundation/drone/db"
	tele "gopkg.in/telebot.v3"
)

const (
	rejectionTag     = "#unfortunately2025"
	bigtechOfferTag  = "#bigtech_offer2025"
	faangOfferTag    = "#faang_offer2025"
	seniorPromoTag   = "#senior_promo2025"
	staffPromoTag    = "#staff_promo2025"
	usaRelocationTag = "#usa2025"

	bigtechOfferOkrGoal  = 5
	faangOfferOkrGoal    = 3
	seniorPromoOkrGoal   = 3
	staffPromoOkrGoal    = 1
	usaRelocationOkrGoal = 1
	rejectionOkrGoal     = 300

	removeCommand = "/remove_okr"
)

var okrTags = []string{
	rejectionTag,
	bigtechOfferTag,
	faangOfferTag,
	seniorPromoTag,
	staffPromoTag,
	usaRelocationTag,
}

type OkrProgress struct {
	Current int
	Goal    int
	Tag     string
	Status  string
}

type OkrTemplateData struct {
	Bigtech   OkrProgress
	Faang     OkrProgress
	Senior    OkrProgress
	Staff     OkrProgress
	Usa       OkrProgress
	Rejection OkrProgress
}

var okrMessageTemplate = template.Must(template.New("okr").Parse(`ОКРы 2025:

Офферы:
{{.Bigtech.Current}}/{{.Bigtech.Goal}} в бигтех ({{.Bigtech.Tag}}) {{.Bigtech.Status}}
{{.Faang.Current}}/{{.Faang.Goal}} в FAANG ({{.Faang.Tag}}) {{.Faang.Status}}

Промо:
{{.Senior.Current}}/{{.Senior.Goal}} на сеньора ({{.Senior.Tag}}) {{.Senior.Status}}
{{.Staff.Current}}/{{.Staff.Goal}} на стаффа ({{.Staff.Tag}}) {{.Staff.Status}}

Релокация:
{{.Usa.Current}}/{{.Usa.Goal}} релока в США ({{.Usa.Tag}}) {{.Usa.Status}}

Анфочантли:
{{.Rejection.Current}}/{{.Rejection.Goal}} ({{.Rejection.Tag}}) {{.Rejection.Status}}`))

func (s *Service) OnUpdateOkr(ctx context.Context, c tele.Context) error {
	msg, chat := c.Message(), c.Chat()
	if msg == nil || chat == nil || chat.ID != s.cfg.ChatID {
		return nil
	}

	text := strings.TrimSpace(strings.ToLower(msg.Text))
	var tagToUpdate string

	for _, tag := range okrTags {
		if strings.Contains(text, tag) {
			tagToUpdate = tag
			break
		}
	}

	if tagToUpdate == "" {
		return nil
	}

	return s.database.Do(ctx, func(tx db.Tx) error {
		okrValues, err := db.GetJsonDefault(tx, keyOkrValues, make(map[string]int))
		if err != nil {
			return fmt.Errorf("get okr values: %w", err)
		}

		toRemove := strings.HasPrefix(text, removeCommand)
		if toRemove {
			if okrValues[tagToUpdate] > 0 {
				okrValues[tagToUpdate]--
			}
		} else {
			okrValues[tagToUpdate]++
		}

		if err := db.SetJson(tx, keyOkrValues, okrValues); err != nil {
			return fmt.Errorf("save okr values: %w", err)
		}

		progressMessage, err := constructOkrProgressMessage(okrValues)
		if err != nil {
			return fmt.Errorf("construct progress message: %w", err)
		}

		// get pinned message ID if exists
		pinnedMessageID, err := db.GetJson[int](tx, keyOkrPinnedMessage)

		// if message hasn't been posted yet, post it, pin it and save the id
		if errors.Is(err, db.ErrKeyNotFound) {
			if err := postNewOkrMessage(s, tx, progressMessage); err != nil {
				return fmt.Errorf("post initial okr message: %w", err)
			}

			return nil
		} else if err != nil {
			return fmt.Errorf("get pinned message id: %w", err)
		}

		_, err = s.telegram.EditMessageText(pinnedMessageID, progressMessage)
		if err != nil {
			return fmt.Errorf("edit pinned message: %w", err)
		}

		return nil
	})
}

func postNewOkrMessage(s *Service, tx db.Tx, progressMessage string) error {
	messageID, err := s.telegram.SendText(s.cfg.InterviewsThreadID, progressMessage)
	if err != nil {
		return fmt.Errorf("send okr message: %w", err)
	}

	if err := s.telegram.Pin(messageID); err != nil {
		return fmt.Errorf("pin okr message: %w", err)
	}

	if err := db.SetJson(tx, keyOkrPinnedMessage, messageID); err != nil {
		return fmt.Errorf("save new pinned message id: %w", err)
	}

	return nil
}

func constructOkrProgressMessage(values map[string]int) (string, error) {
	data := OkrTemplateData{
		Bigtech: OkrProgress{
			Current: values[bigtechOfferTag],
			Goal:    bigtechOfferOkrGoal,
			Tag:     bigtechOfferTag,
			Status:  getStatusEmoji(values[bigtechOfferTag], bigtechOfferOkrGoal),
		},
		Faang: OkrProgress{
			Current: values[faangOfferTag],
			Goal:    faangOfferOkrGoal,
			Tag:     faangOfferTag,
			Status:  getStatusEmoji(values[faangOfferTag], faangOfferOkrGoal),
		},
		Senior: OkrProgress{
			Current: values[seniorPromoTag],
			Goal:    seniorPromoOkrGoal,
			Tag:     seniorPromoTag,
			Status:  getStatusEmoji(values[seniorPromoTag], seniorPromoOkrGoal),
		},
		Staff: OkrProgress{
			Current: values[staffPromoTag],
			Goal:    staffPromoOkrGoal,
			Tag:     staffPromoTag,
			Status:  getStatusEmoji(values[staffPromoTag], staffPromoOkrGoal),
		},
		Usa: OkrProgress{
			Current: values[usaRelocationTag],
			Goal:    usaRelocationOkrGoal,
			Tag:     usaRelocationTag,
			Status:  getStatusEmoji(values[usaRelocationTag], usaRelocationOkrGoal),
		},
		Rejection: OkrProgress{
			Current: values[rejectionTag],
			Goal:    rejectionOkrGoal,
			Tag:     rejectionTag,
			Status:  getStatusEmoji(values[rejectionTag], rejectionOkrGoal),
		},
	}

	var buf bytes.Buffer
	if err := okrMessageTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute okr template: %w", err)
	}
	return buf.String(), nil
}

func getStatusEmoji(value, goal int) string {
	if value >= goal {
		return "✅"
	}
	return "🔄"
}
