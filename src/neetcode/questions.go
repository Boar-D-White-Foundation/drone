package neetcode

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed questions.json
var rawQuestions []byte

type Question struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Slug       string `json:"title_slug"`
	Difficulty string `json:"difficulty"`
}

func (q Question) LeetcodeLink() string {
	return fmt.Sprintf("https://leetcode.com/problems/%s/", q.Slug)
}

func (q Question) LeetcodeCaLink() string {
	return fmt.Sprintf("https://leetcode.ca/all/%d.html", q.ID)
}

func Questions() ([]Question, error) {
	var qs []Question
	err := json.Unmarshal(rawQuestions, &qs)
	return qs, err
}

func QuestionsByDifficulty(difficulty string) ([]Question, error) {
	qs, err := Questions()
	if err != nil {
		return nil, err
	}

	var filtered []Question
	for _, q := range qs {
		if q.Difficulty == difficulty {
			filtered = append(filtered, q)
		}
	}
	return filtered, nil
}
