package tg

var (
	ReactionClown    = NewReactionEmoji("🤡")
	ReactionOk       = NewReactionEmoji("👌")
	ReactionMoai     = NewReactionEmoji("🗿")
	ReactionEgor     = NewReactionEmoji("🌚")
	ReactionThumbsUp = NewReactionEmoji("👍")
	ReactionHotDog   = NewReactionEmoji("🌭")
)

type Reaction struct {
	Type    string `json:"type"`
	Emoji   string `json:"emoji,omitempty"`
	EmojiID string `json:"custom_emoji_id,omitempty"`
}

func NewReactionEmoji(emoji string) Reaction {
	return Reaction{
		Type:  "emoji",
		Emoji: emoji,
	}
}

func NewReactionCustomEmoji(emojiID string) Reaction {
	return Reaction{
		Type:    "custom_emoji",
		EmojiID: emojiID,
	}
}

func (e Reaction) IsCustom() bool {
	return e.Type == "custom_emoji"
}
