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
)

type TranslateHandler struct {
	client *http.Client
}

func NewTranslateHandler() *TranslateHandler {
	return &TranslateHandler{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

type translateRequest struct {
	Text   string `json:"text" binding:"required"`
	Target string `json:"target" binding:"required"`
	Source string `json:"source"`
}

type translateResponse struct {
	TranslatedText string `json:"translated_text"`
}

type libreRequest struct {
	Q      string `json:"q"`
	Source string `json:"source"`
	Target string `json:"target"`
	Format string `json:"format"`
}

func (h *TranslateHandler) Translate(c *gin.Context) {
	var req translateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	text := strings.TrimSpace(req.Text)
	if len(text) > 2000 {
		text = text[:2000]
	}
	if text == "" {
		c.JSON(http.StatusOK, gin.H{"data": ""})
		return
	}

	source := req.Source
	if source == "" || source == "auto" {
		source = "en"
	}
	target := strings.Split(req.Target, "-")[0]

	body, _ := json.Marshal(libreRequest{
		Q:      text,
		Source: source,
		Target: target,
		Format: "text",
	})

	resp, err := h.client.Post("https://libretranslate.com/translate", "application/json", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("translate service unavailable: %v", err)})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("translate returned %d", resp.StatusCode)})
		return
	}

	var result translateResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse translate response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result.TranslatedText})
}
