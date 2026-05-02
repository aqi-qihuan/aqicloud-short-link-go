package trpc

import (
	"context"

	"github.com/aqi/aqicloud-short-link-go/internal/ai/agent"
	"github.com/aqi/aqicloud-short-link-go/internal/ai/config"
)

// AgentService wraps AI agents as a trpc-agent-go compatible service.
// This allows the AI features to be exposed as RPC endpoints via trpc-agent-go framework.
type AgentService struct {
	recommend *agent.RecommendationAgent
	analytics *agent.AnalyticsAgent
	safety    *agent.URLSafetyAgent
}

func NewAgentService(cfg *config.AIConfig) *AgentService {
	return &AgentService{
		recommend: agent.NewRecommendationAgent(cfg.APIKey, cfg.BaseURL, cfg.ModelName),
		analytics: agent.NewAnalyticsAgent(cfg.APIKey, cfg.BaseURL, cfg.ModelName),
		safety:    agent.NewURLSafetyAgent(cfg.APIKey, cfg.BaseURL, cfg.ModelName),
	}
}

// RecommendRequest is the RPC request for recommendation.
type RecommendRequest struct {
	URL string `json:"url"`
}

// RecommendResponse is the RPC response for recommendation.
type RecommendResponse struct {
	Title       string   `json:"title"`
	GroupSuggest string  `json:"group_suggest"`
	Tags        []string `json:"tags"`
	Summary     string   `json:"summary"`
}

// Recommend provides smart short link recommendations via RPC.
func (s *AgentService) Recommend(ctx context.Context, req *RecommendRequest) (*RecommendResponse, error) {
	result, err := s.recommend.Recommend(ctx, req.URL)
	if err != nil {
		return nil, err
	}
	return &RecommendResponse{
		Title:       result.Title,
		GroupSuggest: result.GroupSuggest,
		Tags:        result.Tags,
		Summary:     result.Summary,
	}, nil
}

// AnalyticsRequest is the RPC request for analytics.
type AnalyticsRequest struct {
	Question      string `json:"question"`
	ShortLinkCode string `json:"shortLinkCode"`
}

// AnalyticsResponse is the RPC response for analytics.
type AnalyticsResponse struct {
	SQL         string `json:"sql"`
	Explanation string `json:"explanation"`
}

// Analytics converts natural language to ClickHouse SQL via RPC.
func (s *AgentService) Analytics(ctx context.Context, req *AnalyticsRequest) (*AnalyticsResponse, error) {
	result, err := s.analytics.Query(ctx, req.Question, req.ShortLinkCode)
	if err != nil {
		return nil, err
	}
	return &AnalyticsResponse{
		SQL:         result.SQL,
		Explanation: result.Explanation,
	}, nil
}

// SafetyCheckRequest is the RPC request for URL safety check.
type SafetyCheckRequest struct {
	URL string `json:"url"`
}

// SafetyCheckResponse is the RPC response for URL safety check.
type SafetyCheckResponse struct {
	Safe   bool     `json:"safe"`
	Reason string   `json:"reason,omitempty"`
	Tags   []string `json:"tags,omitempty"`
	Score  float64  `json:"score"`
}

// CheckSafety checks URL safety via RPC.
func (s *AgentService) CheckSafety(ctx context.Context, req *SafetyCheckRequest) (*SafetyCheckResponse, error) {
	result, err := s.safety.Analyze(ctx, req.URL)
	if err != nil {
		return nil, err
	}
	return &SafetyCheckResponse{
		Safe:   result.Safe,
		Reason: result.Reason,
		Tags:   result.Tags,
		Score:  result.Score,
	}, nil
}
