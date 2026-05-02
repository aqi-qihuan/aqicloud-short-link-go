package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aqi/aqicloud-short-link-go/internal/ai/llm"
)

// RecommendationAgent provides smart short link recommendations.
type RecommendationAgent struct {
	client *llm.Client
}

func NewRecommendationAgent(apiKey, baseURL, model string) *RecommendationAgent {
	return &RecommendationAgent{
		client: llm.NewClient(baseURL, apiKey, model),
	}
}

// RecommendationResult holds the AI-generated suggestions.
type RecommendationResult struct {
	Title        string   `json:"title"`
	GroupSuggest string   `json:"group_suggest"`
	Tags         []string `json:"tags"`
	Summary      string   `json:"summary"`
}

// Recommend analyzes a URL and returns suggested title/group/tags.
func (a *RecommendationAgent) Recommend(ctx context.Context, url string) (*RecommendationResult, error) {
	systemPrompt := `You are a URL analysis assistant. Analyze the given URL and return a JSON object with these fields:
- "title": a concise Chinese title (max 20 chars)
- "group_suggest": a suggested group/category name in Chinese
- "tags": an array of 2-4 relevant tags in Chinese
- "summary": a one-sentence Chinese summary of what the URL links to

Return ONLY valid JSON, no other text.`

	messages := []llm.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("Analyze this URL: %s", url)},
	}

	resp, err := a.client.Chat(messages, 512, 0.7)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	jsonStr := llm.ExtractJSON(resp)

	var result RecommendationResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		// Fallback: return raw response as summary
		return &RecommendationResult{
			Title:   "",
			Summary: resp,
		}, nil
	}
	return &result, nil
}
