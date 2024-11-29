package boardwhite

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	tele "gopkg.in/telebot.v3"
)

var (
	twitterLinkRe = regexp.MustCompile(`(http[s]?:\/\/)?(www\.)?\bx\.com\/[-a-zA-Z0-9@:%_\+~#?&\/\/=]+`)
)

const (
	twitterDomainName = "x.com"
	fixupxDomainName  = "i.fixupx.com"
)

func (s *Service) OnPostTwitterEmbed(ctx context.Context, c tele.Context) error {
	msg, chat := c.Message(), c.Chat()
	if msg == nil || chat == nil || chat.ID != s.cfg.ChatID {
		return nil
	}

	matchedTwitterLinks := twitterLinkRe.FindStringSubmatch(msg.Text)
	if len(matchedTwitterLinks) == 0 {
		return nil
	}

	firstTwitterLink := matchedTwitterLinks[0]
	embedTwitterLink := strings.Replace(firstTwitterLink, twitterDomainName, fixupxDomainName, 1)

	_, err := s.telegram.ReplyWithText(msg.ID, embedTwitterLink)
	if err != nil {
		return fmt.Errorf("reply with twitter embed: %w", err)
	}

	return nil
}
