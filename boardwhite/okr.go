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

type okrProgress struct {
	Current int
	Goal    int
	Tag     string
	Status  string
}

type okrTemplateData struct {
	Bigtech   okrProgress
	Faang     okrProgress
	Senior    okrProgress
	Staff     okrProgress
	Usa       okrProgress
	Rejection okrProgress
}

type okrs struct {
	Values  map[string]int         `json:"values"`  // from tag to total value of this tag
	Updates map[string][]okrUpdate `json:"updates"` // from tag to list of updates
}

// one update might contain same tag multiple times,
// so we need to keep track what value of the praticular okr tag
// is bound to the particular update
type okrUpdate struct {
	Update tele.Update `json:"update"`
	Value  int         `json:"value"`
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
	msg, chat, update := c.Message(), c.Chat(), c.Update()
	if msg == nil || chat == nil || chat.ID != s.cfg.ChatID || msg.IsForwarded() {
		return nil
	}

	tagsToUpdate := extractOkrTags(msg.Text)
	if len(tagsToUpdate) == 0 {
		return nil
	}

	if strings.HasPrefix(msg.Text, removeCommand) {
		return nil
	}

	return s.database.Do(ctx, func(tx db.Tx) error {
		okrs, err := db.GetJsonDefault(tx, keyOkrValues, okrs{})
		if err != nil {
			return fmt.Errorf("get okr values: %w", err)
		}

		if okrs.Values == nil {
			okrs.Values = make(map[string]int)
		}

		if okrs.Updates == nil {
			okrs.Updates = make(map[string][]okrUpdate)
		}

		for tagToUpdate, updateValue := range tagsToUpdate {
			okrs.Values[tagToUpdate] += updateValue

			okrUpdate := okrUpdate{Update: update, Value: updateValue}
			okrs.Updates[tagToUpdate] = append(okrs.Updates[tagToUpdate], okrUpdate)
		}

		if err := db.SetJson(tx, keyOkrValues, okrs); err != nil {
			return fmt.Errorf("save okr values: %w", err)
		}

		progressMessage, err := constructOkrProgressMessage(okrs.Values)
		if err != nil {
			return fmt.Errorf("construct progress message: %w", err)
		}

		return s.updatePinnedOkrMessage(tx, progressMessage)
	})
}

func (s *Service) OnRemoveOkr(ctx context.Context, c tele.Context) error {
	msg, chat := c.Message(), c.Chat()
	if msg == nil || chat == nil || chat.ID != s.cfg.ChatID {
		return nil
	}

	if !strings.HasPrefix(msg.Text, removeCommand) {
		return nil
	}

	if !msg.IsReply() {
		return nil
	}

	originalOkrMessage := msg.ReplyTo

	tagsToRemoveFrom := extractOkrTags(msg.Text)
	originalTags := extractOkrTags(originalOkrMessage.Text)

	// remove all tags in from the original message
	if msg.Text == removeCommand {
		tagsToRemoveFrom = originalTags
	}

	// prevent removing more than was in the original message when removing specific tags
	for tag, removeValue := range tagsToRemoveFrom {
		originalTagValue := originalTags[tag]
		tagsToRemoveFrom[tag] = min(removeValue, originalTagValue)
	}

	return s.database.Do(ctx, func(tx db.Tx) error {
		okrs, err := db.GetJsonDefault(tx, keyOkrValues, okrs{})
		if err != nil {
			return fmt.Errorf("get okr values: %w", err)
		}

		if okrs.Values == nil {
			okrs.Values = make(map[string]int)
		}

		if okrs.Updates == nil {
			okrs.Updates = make(map[string][]okrUpdate)
		}

		isUpdated := false
		for tagToRemoveFrom, removeValue := range tagsToRemoveFrom {
			if removeValue == 0 {
				continue
			}

			okrUpdates := okrs.Updates[tagToRemoveFrom]

			for i, update := range okrUpdates {

				if update.Update.Message.ID == originalOkrMessage.ID {

					newUpdateValue := update.Value - removeValue

					if newUpdateValue <= 0 {
						okrUpdates[i] = okrUpdates[len(okrUpdates)-1]
						okrUpdates = okrUpdates[:len(okrUpdates)-1]

						okrs.Updates[tagToRemoveFrom] = okrUpdates
					} else {
						okrUpdates[i].Value = newUpdateValue
					}

					if okrs.Values[tagToRemoveFrom] > 0 {
						okrs.Values[tagToRemoveFrom] -= removeValue
					}

					isUpdated = true

				}
			}
		}

		if isUpdated {
			progressMessage, err := constructOkrProgressMessage(okrs.Values)
			if err != nil {
				return fmt.Errorf("construct progress message: %w", err)
			}

			if err := s.updatePinnedOkrMessage(tx, progressMessage); err != nil {
				return fmt.Errorf("udpate pinned message: %w", err)
			}

			if err := db.SetJson(tx, keyOkrValues, okrs); err != nil {
				return fmt.Errorf("save okr values: %w", err)
			}
		}

		return nil
	})
}

func extractOkrTag(msgText string) string {
	for _, tag := range okrTags {
		if strings.Contains(msgText, tag) {
			return tag
		}
	}

	return ""
}

// returns map from tag to number of its occurances in the text
func extractOkrTags(msgText string) map[string]int {
	tags := make(map[string]int)
	for _, tag := range okrTags {
		occurances := strings.Count(msgText, tag)

		if occurances > 0 {
			tags[tag] = occurances
		}
	}

	return tags
}

func (s *Service) updatePinnedOkrMessage(tx db.Tx, progressMessage string) error {
	// get pinned message ID if exists
	pinnedMessageID, err := db.GetJson[int](tx, keyOkrPinnedMessage)

	// if message hasn't been posted yet, post it, pin it and save the id
	if errors.Is(err, db.ErrKeyNotFound) {
		if err := s.postNewOkrMessage(tx, progressMessage); err != nil {
			return fmt.Errorf("post initial okr message: %w", err)
		}

		return nil
	}

	if err != nil {
		return fmt.Errorf("get pinned message id: %w", err)
	}

	_, err = s.telegram.EditMessageText(pinnedMessageID, progressMessage)
	if err != nil {
		return fmt.Errorf("edit pinned message: %w", err)
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
		return fmt.Errorf("save new pinned message id: %w", err)
	}

	return nil
}

func constructOkrProgressMessage(values map[string]int) (string, error) {
	data := okrTemplateData{
		Bigtech: okrProgress{
			Current: values[bigtechOfferTag],
			Goal:    bigtechOfferOkrGoal,
			Tag:     bigtechOfferTag,
			Status:  getStatusEmoji(values[bigtechOfferTag], bigtechOfferOkrGoal),
		},
		Faang: okrProgress{
			Current: values[faangOfferTag],
			Goal:    faangOfferOkrGoal,
			Tag:     faangOfferTag,
			Status:  getStatusEmoji(values[faangOfferTag], faangOfferOkrGoal),
		},
		Senior: okrProgress{
			Current: values[seniorPromoTag],
			Goal:    seniorPromoOkrGoal,
			Tag:     seniorPromoTag,
			Status:  getStatusEmoji(values[seniorPromoTag], seniorPromoOkrGoal),
		},
		Staff: okrProgress{
			Current: values[staffPromoTag],
			Goal:    staffPromoOkrGoal,
			Tag:     staffPromoTag,
			Status:  getStatusEmoji(values[staffPromoTag], staffPromoOkrGoal),
		},
		Usa: okrProgress{
			Current: values[usaRelocationTag],
			Goal:    usaRelocationOkrGoal,
			Tag:     usaRelocationTag,
			Status:  getStatusEmoji(values[usaRelocationTag], usaRelocationOkrGoal),
		},
		Rejection: okrProgress{
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
