package handler

import (
	"github.com/aqi/aqicloud-short-link-go/internal/ai/agent"
	"github.com/aqi/aqicloud-short-link-go/internal/ai/config"
	"github.com/aqi/aqicloud-short-link-go/internal/common/interceptor"
	"github.com/aqi/aqicloud-short-link-go/internal/common/response"
	"github.com/gin-gonic/gin"
)

// AIHandler exposes AI features as HTTP endpoints.
type AIHandler struct {
	recommend *agent.RecommendationAgent
	analytics *agent.AnalyticsAgent
	safety    *agent.URLSafetyAgent
}

func NewAIHandler(cfg *config.AIConfig) *AIHandler {
	return &AIHandler{
		recommend: agent.NewRecommendationAgent(cfg.APIKey, cfg.BaseURL, cfg.ModelName),
		analytics: agent.NewAnalyticsAgent(cfg.APIKey, cfg.BaseURL, cfg.ModelName),
		safety:    agent.NewURLSafetyAgent(cfg.APIKey, cfg.BaseURL, cfg.ModelName),
	}
}

// Recommend handles POST /api/ai/v1/recommend
func (h *AIHandler) Recommend(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildError("login required"))
		return
	}

	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, response.BuildError("url is required"))
		return
	}

	result, err := h.recommend.Recommend(c.Request.Context(), req.URL)
	if err != nil {
		response.JSON(c, response.BuildError("recommendation failed"))
		return
	}
	response.JSON(c, response.BuildSuccessData(result))
}

// Analytics handles POST /api/ai/v1/analytics
func (h *AIHandler) Analytics(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildError("login required"))
		return
	}

	var req struct {
		Question      string `json:"question" binding:"required"`
		ShortLinkCode string `json:"shortLinkCode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, response.BuildError("question is required"))
		return
	}

	result, err := h.analytics.Query(c.Request.Context(), req.Question, req.ShortLinkCode)
	if err != nil {
		response.JSON(c, response.BuildError("analytics failed"))
		return
	}
	response.JSON(c, response.BuildSuccessData(result))
}

// CheckSafety handles POST /api/ai/v1/check_safety
func (h *AIHandler) CheckSafety(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildError("login required"))
		return
	}

	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, response.BuildError("url is required"))
		return
	}

	result, err := h.safety.Analyze(c.Request.Context(), req.URL)
	if err != nil {
		response.JSON(c, response.BuildError("safety check failed"))
		return
	}
	response.JSON(c, response.BuildSuccessData(result))
}
