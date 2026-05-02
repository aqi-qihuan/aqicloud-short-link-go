package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aqi/aqicloud-short-link-go/internal/ai/llm"
)

// AnalyticsAgent converts natural language questions into ClickHouse SQL queries.
type AnalyticsAgent struct {
	client *llm.Client
}

func NewAnalyticsAgent(apiKey, baseURL, model string) *AnalyticsAgent {
	return &AnalyticsAgent{
		client: llm.NewClient(baseURL, apiKey, model),
	}
}

// AnalyticsResult holds the generated SQL and explanation.
type AnalyticsResult struct {
	SQL         string `json:"sql"`
	Explanation string `json:"explanation"`
}

const clickHouseSchema = `Table: visit_stats (ClickHouse MergeTree)
Columns:
  code          String    -- 短链码
  referer       String    -- 来源页面
  is_new        String    -- 是否新访客 ('0'/'1')
  account_no    UInt64    -- 账号编号
  province      String    -- 省份
  city          String    -- 城市
  ip            String    -- 访客IP
  browser_name  String    -- 浏览器
  os            String    -- 操作系统
  device_type   String    -- 设备类型 (PC/Mobile/Tablet)
  pv            UInt64    -- 页面浏览量
  uv            UInt64    -- 独立访客数
  start_time    DateTime  -- 访问开始时间
  end_time      DateTime  -- 访问结束时间
  ts            UInt64    -- 访问时间戳(毫秒)

ClickHouse specific functions: toYYYYMMDD(), toHour(), toMinute(), toYYYYMMDDhhmmss()`

// Query converts a natural language question into a ClickHouse SQL query.
func (a *AnalyticsAgent) Query(ctx context.Context, question string, shortLinkCode string) (*AnalyticsResult, error) {
	systemPrompt := fmt.Sprintf(`You are a ClickHouse SQL expert. Given the table schema below, generate a SQL query to answer the user's question.

%s

Rules:
1. Always filter by account_no and code when a short link code is provided.
2. Use ClickHouse-specific functions (toYYYYMMDD, toHour, toMinute) for date operations.
3. Return a JSON object with "sql" and "explanation" fields.
4. "explanation" should be a brief Chinese description of what the query does.
5. Return ONLY valid JSON, no other text.

Example response:
{"sql": "SELECT count() FROM visit_stats WHERE code = 'abc'", "explanation": "查询短链abc的总访问量"}`, clickHouseSchema)

	userMsg := fmt.Sprintf("Question: %s", question)
	if shortLinkCode != "" {
		userMsg = fmt.Sprintf("Short link code: %s\nQuestion: %s", shortLinkCode, question)
	}

	messages := []llm.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMsg},
	}

	resp, err := a.client.Chat(messages, 1024, 0.3)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	jsonStr := llm.ExtractJSON(resp)

	var result AnalyticsResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		// Fallback: return raw response as explanation
		return &AnalyticsResult{
			Explanation: resp,
		}, nil
	}
	return &result, nil
}
