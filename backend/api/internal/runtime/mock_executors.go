package runtime

import (
	"context"
	"sort"
	"strings"
	"unicode"
)

type EchoExecutor struct{}

func (EchoExecutor) Run(_ context.Context, input map[string]any, _ map[string]any) (map[string]any, error) {
	return map[string]any{
		"echo": input,
	}, nil
}

type TextSummaryMockExecutor struct{}

func (TextSummaryMockExecutor) Run(_ context.Context, input map[string]any, config map[string]any) (map[string]any, error) {
	text := strings.TrimSpace(asString(input["text"]))
	if text == "" {
		text = "No text provided."
	}

	maxLength := 160
	if configured, ok := asInt(config["maxLength"]); ok && configured > 0 {
		maxLength = configured
	}

	summary := text
	if len(summary) > maxLength {
		summary = strings.TrimSpace(summary[:maxLength]) + "..."
	}

	return map[string]any{
		"result": summary,
	}, nil
}

type KeywordExtractMockExecutor struct{}

func (KeywordExtractMockExecutor) Run(_ context.Context, input map[string]any, config map[string]any) (map[string]any, error) {
	text := strings.ToLower(asString(input["text"]))
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	minLength := 4
	if configured, ok := asInt(config["minLength"]); ok && configured > 0 {
		minLength = configured
	}

	frequencies := make(map[string]int)
	for _, word := range words {
		if len(word) < minLength {
			continue
		}
		frequencies[word]++
	}

	type keywordCount struct {
		word  string
		count int
	}

	keywordCounts := make([]keywordCount, 0, len(frequencies))
	for word, count := range frequencies {
		keywordCounts = append(keywordCounts, keywordCount{word: word, count: count})
	}

	sort.Slice(keywordCounts, func(i, j int) bool {
		if keywordCounts[i].count == keywordCounts[j].count {
			return keywordCounts[i].word < keywordCounts[j].word
		}
		return keywordCounts[i].count > keywordCounts[j].count
	})

	limit := 5
	if configured, ok := asInt(config["limit"]); ok && configured > 0 {
		limit = configured
	}
	if limit > len(keywordCounts) {
		limit = len(keywordCounts)
	}

	keywords := make([]string, 0, limit)
	for _, item := range keywordCounts[:limit] {
		keywords = append(keywords, item.word)
	}

	return map[string]any{
		"keywords": keywords,
	}, nil
}

func DefaultRegistry() *Registry {
	return NewRegistry(map[string]Executor{
		"echo":                 EchoExecutor{},
		"text-summary-mock":    TextSummaryMockExecutor{},
		"keyword-extract-mock": KeywordExtractMockExecutor{},
	})
}

func asString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}

func asInt(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case float64:
		return int(typed), true
	default:
		return 0, false
	}
}
