package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aqi/aqicloud-short-link-go/internal/ai/llm"
)

// URLSafetyAgent detects malicious URLs before creating short links.
type URLSafetyAgent struct {
	client *llm.Client
}

func NewURLSafetyAgent(apiKey, baseURL, model string) *URLSafetyAgent {
	return &URLSafetyAgent{
		client: llm.NewClient(baseURL, apiKey, model),
	}
}

// SafetyResult holds the URL safety analysis.
type SafetyResult struct {
	Safe   bool     `json:"safe"`
	Reason string   `json:"reason,omitempty"`
	Tags   []string `json:"tags,omitempty"` // phishing, malware, scam, gambling, etc.
	Score  float64  `json:"score"`          // 0.0 (safe) to 1.0 (dangerous)
}

// Analyze checks if a URL is potentially malicious.
func (a *URLSafetyAgent) Analyze(ctx context.Context, url string) (*SafetyResult, error) {
	systemPrompt := `You are a URL security analyst. Analyze the given URL for safety concerns.
Check for: phishing, malware distribution, scam patterns, gambling, adult content, fraud.

Return a JSON object with:
- "safe": boolean (true if the URL appears safe)
- "reason": brief Chinese explanation of the verdict
- "tags": array of risk category tags (e.g., "钓鱼", "恶意软件", "诈骗", "赌博")
- "score": risk score from 0.0 (completely safe) to 1.0 (extremely dangerous)

Return ONLY valid JSON, no other text.`

	messages := []llm.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("Analyze this URL: %s", url)},
	}

	resp, err := a.client.Chat(messages, 512, 0.3)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	jsonStr := llm.ExtractJSON(resp)

	var result SafetyResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		// Fallback: assume safe if LLM response can't be parsed
		return &SafetyResult{
			Safe:   true,
			Score:  0.0,
			Reason: resp,
		}, nil
	}
	return &result, nil
}
