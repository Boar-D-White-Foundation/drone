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
