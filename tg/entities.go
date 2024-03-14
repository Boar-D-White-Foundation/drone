package tg

type ReactionTypeEmoji struct {
	Emoji string `json:"emoji,omitempty"`
}

type ReactionTypeCustomEmoji struct {
	EmojiID string `json:"custom_emoji_id,omitempty"`
}

type Emoji struct {
	Type string `json:"type"`
	ReactionTypeEmoji
	ReactionTypeCustomEmoji
}

type ReactionEmoji struct {
	Type  string `json:"type"`
	Emoji string `json:"emoji"`
}

type ReactionOptions struct {
	MessageID int              `json:"message_id"`
	ChatID    int64            `json:"chat_id"`
	Reaction  []*ReactionEmoji `json:"reaction"`
	IsBig     bool             `json:"is_big"`
}

var (
	ReactionClown    = &ReactionEmoji{Type: "emoji", Emoji: "🤡"}
	ReactionEgor     = &ReactionEmoji{Type: "emoji", Emoji: "🌚"}
	ReactionThumbsUp = &ReactionEmoji{Type: "emoji", Emoji: "👍"}
	ReactionHotDog   = &ReactionEmoji{Type: "emoji", Emoji: "🌭"}
)
