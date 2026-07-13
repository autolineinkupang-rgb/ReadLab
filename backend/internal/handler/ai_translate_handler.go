package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
)

type AITranslateHandler struct {
	DB      *gorm.DB
	client  *http.Client
}

func NewAITranslateHandler(db *gorm.DB) *AITranslateHandler {
	return &AITranslateHandler{
		DB:     db,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

type aiTranslateSettings struct {
	Provider          string `json:"provider"`
	Model             string `json:"model"`
	Endpoint          string `json:"endpoint"`
	Key               string `json:"key,omitempty"`
	TargetLanguage    string `json:"target_language"`
	Instruction       string `json:"instruction"`
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

func (h *AITranslateHandler) GetSettings(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user model.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"provider":         user.AITranslateProvider,
		"model":            user.AITranslateModel,
		"endpoint":         user.AITranslateEndpoint,
		"key":              maskKey(user.AITranslateKey),
		"has_key":          user.AITranslateKey != "",
		"target_language":  user.TranslateTargetLang,
		"instruction":      user.AITranslateInstruction,
	})
}

func (h *AITranslateHandler) UpdateSettings(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req aiTranslateSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"ai_translate_provider":   req.Provider,
		"ai_translate_model":      req.Model,
		"ai_translate_endpoint":   req.Endpoint,
		"translate_target_lang":   req.TargetLanguage,
		"ai_translate_instruction": req.Instruction,
	}
	if req.Key != "" {
		updates["ai_translate_key"] = req.Key
	}

	if err := h.DB.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "settings updated"})
}

type aiTranslateRequest struct {
	Text   string `json:"text" binding:"required"`
	Target string `json:"target,omitempty"`
	Source string `json:"source,omitempty"`
}

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterRequest struct {
	Model    string              `json:"model"`
	Messages []openRouterMessage `json:"messages"`
}

type openRouterChoice struct {
	Message openRouterMessage `json:"message"`
}

type openRouterResponse struct {
	Choices []openRouterChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (h *AITranslateHandler) Translate(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user model.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if user.AITranslateKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "AI translate key not configured. Set it in your profile settings."})
		return
	}

	var req aiTranslateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	text := strings.TrimSpace(req.Text)
	if len(text) > 5000 {
		text = text[:5000]
	}
	if text == "" {
		c.JSON(http.StatusOK, gin.H{"data": ""})
		return
	}

	targetCode := req.Target
	if targetCode == "" {
		targetCode = user.TranslateTargetLang
	}
	target := strings.Split(targetCode, "-")[0]
	targetLang := targetToLanguage(target)

	messages := []openRouterMessage{}
	if user.AITranslateInstruction != "" {
		messages = append(messages, openRouterMessage{Role: "system", Content: user.AITranslateInstruction})
	}
	messages = append(messages, openRouterMessage{
		Role:    "user",
		Content: fmt.Sprintf("Translate the following text to %s. Return only the translated text, nothing else.\n\nText: %s", targetLang, text),
	})

	body, _ := json.Marshal(openRouterRequest{
		Model:    user.AITranslateModel,
		Messages: messages,
	})

	httpReq, _ := http.NewRequest("POST", user.AITranslateEndpoint, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+user.AITranslateKey)

	resp, err := h.client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("AI translate service unavailable: %v", err)})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result openRouterResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse AI translate response"})
		return
	}

	if result.Error != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("AI translate error: %s", result.Error.Message)})
		return
	}

	if len(result.Choices) == 0 {
		c.JSON(http.StatusOK, gin.H{"data": text})
		return
	}

	translated := strings.TrimSpace(result.Choices[0].Message.Content)
	c.JSON(http.StatusOK, gin.H{"data": translated})
}

func targetToLanguage(code string) string {
	langs := map[string]string{
		"id": "Indonesian",
		"en": "English",
		"ja": "Japanese",
		"ko": "Korean",
		"zh": "Chinese",
		"fr": "French",
		"de": "German",
		"es": "Spanish",
		"pt": "Portuguese",
		"ru": "Russian",
		"ar": "Arabic",
		"hi": "Hindi",
		"th": "Thai",
		"vi": "Vietnamese",
	}
	if name, ok := langs[code]; ok {
		return name
	}
	return code
}
