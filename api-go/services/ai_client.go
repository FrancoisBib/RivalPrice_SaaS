package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

// AIClient calls OpenAI API to generate summaries and recommendations
type AIClient struct {
	apiKey     string
	httpClient *http.Client
	model      string
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature"`
}

type openAIChoice struct {
	Message openAIMessage `json:"message"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
}

// AIInsight contains AI-generated analysis
type AIInsight struct {
	Summary        string
	Recommendation string
	ImpactLevel    int    // 1-10 score
	Model          string
}

func NewAIClient() *AIClient {
	return &AIClient{
		apiKey: os.Getenv("OPENAI_API_KEY"),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		model: "gpt-4o-mini",
	}
}

// Analyze generates a summary, recommendation and impact level for a detected change
func (c *AIClient) Analyze(changeType, pageType, oldPrice, newPrice string, changePercent float64, featuresAdded, featuresRemoved, oldText, newText string) (*AIInsight, error) {
	if c.apiKey == "" {
		log.Println("⚠️  OPENAI_API_KEY not set, using rule-based analysis")
		return c.ruleBasedAnalysis(changeType, oldPrice, newPrice, changePercent), nil
	}

	prompt := fmt.Sprintf(`Analyze this competitor change and respond with JSON only.

Change Type: %s
Page Type: %s
Old Price: %s → New Price: %s (%.1f%%)
Features Added: %s
Features Removed: %s
Old Text: %s
New Text: %s

Respond ONLY with this JSON (impact_level must be 1-10):
{"summary": "1-2 sentence summary", "recommendation": "short actionable recommendation", "impact_level": 7}`,
		changeType, pageType, oldPrice, newPrice, changePercent,
		featuresAdded, featuresRemoved, oldText, newText)

	reqBody := openAIRequest{
		Model: c.model,
		Messages: []openAIMessage{
			{Role: "system", Content: "You are a competitive pricing analyst. Be concise and actionable. Return impact_level as integer 1-10."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   200,
		Temperature: 0.5,
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return c.ruleBasedAnalysis(changeType, oldPrice, newPrice, changePercent), nil
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("⚠️  OpenAI API error: %v, falling back to rule-based", err)
		return c.ruleBasedAnalysis(changeType, oldPrice, newPrice, changePercent), nil
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var openAIResp openAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil || len(openAIResp.Choices) == 0 {
		return c.ruleBasedAnalysis(changeType, oldPrice, newPrice, changePercent), nil
	}

	content := openAIResp.Choices[0].Message.Content

	// Parse JSON from response
	var result map[string]interface{}
	// Extract JSON block if wrapped in markdown
	if idx := strings.Index(content, "{"); idx >= 0 {
		if end := strings.LastIndex(content, "}"); end > idx {
			content = content[idx : end+1]
		}
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return c.ruleBasedAnalysis(changeType, oldPrice, newPrice, changePercent), nil
	}

	// Extract impact_level (default to 5 if not present or invalid)
	impactLevel := 5
	if il, ok := result["impact_level"].(float64); ok {
		impactLevel = int(il)
		if impactLevel < 1 {
			impactLevel = 1
		} else if impactLevel > 10 {
			impactLevel = 10
		}
	}

	summary, _ := result["summary"].(string)
	recommendation, _ := result["recommendation"].(string)

	return &AIInsight{
		Summary:        summary,
		Recommendation: recommendation,
		ImpactLevel:    impactLevel,
		Model:          c.model,
	}, nil
}

// ruleBasedAnalysis generates insights without LLM
func (c *AIClient) ruleBasedAnalysis(changeType, oldPrice, newPrice string, changePercent float64) *AIInsight {
	var summary, recommendation string
	impactLevel := 5

	switch {
	case strings.Contains(changeType, "price_increase"):
		summary = fmt.Sprintf("Competitor raised price from %s to %s (+%.1f%%)", oldPrice, newPrice, changePercent)
		recommendation = "Opportunity to capture price-sensitive customers"
		abs := math.Abs(changePercent)
		if abs >= 30 {
			impactLevel = 9
		} else if abs >= 15 {
			impactLevel = 7
		} else if abs >= 5 {
			impactLevel = 5
		} else {
			impactLevel = 3
		}
	case strings.Contains(changeType, "price_decrease"):
		summary = fmt.Sprintf("Competitor lowered price from %s to %s (%.1f%%)", oldPrice, newPrice, changePercent)
		recommendation = "Consider matching price or reinforcing value proposition"
		abs := math.Abs(changePercent)
		if abs >= 30 {
			impactLevel = 10
		} else if abs >= 15 {
			impactLevel = 8
		} else if abs >= 5 {
			impactLevel = 6
		} else {
			impactLevel = 4
		}
	case strings.Contains(changeType, "feature_added"):
		summary = "Competitor added new features to their offering"
		recommendation = "Evaluate feature gap and update roadmap"
		impactLevel = 7
	case strings.Contains(changeType, "feature_removed"):
		summary = "Competitor removed features from their offering"
		recommendation = "Highlight your superior feature set in marketing"
		impactLevel = 4
	case strings.Contains(changeType, "messaging_change"):
		summary = "Competitor updated their messaging and positioning"
		recommendation = "Review your own messaging for competitive differentiation"
		impactLevel = 5
	default:
		summary = fmt.Sprintf("Competitor made changes: %s", changeType)
		recommendation = "Monitor closely for further changes"
		impactLevel = 3
	}

	return &AIInsight{
		Summary:        summary,
		Recommendation: recommendation,
		ImpactLevel:    impactLevel,
		Model:          "rule-based",
	}
}
