package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func ThreadsListHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: query Postgres
	JSON(w, http.StatusOK, map[string]any{
		"threads": []any{},
	})
}

func ThreadDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Path params with stdlib (Go 1.22+):
	id := r.PathValue("id")
	// TODO: fetch thread + summary by id
	JSON(w, http.StatusOK, map[string]any{
		"id":      id,
		"subject": "stub",
	})
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: read body {query:string}, do embedding + pgvector KNN
	JSON(w, http.StatusOK, map[string]any{
		"results": []any{},
	})
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: read {question}, retrieve K docs, call LLM, return answer + citations
	JSON(w, http.StatusOK, map[string]any{
		"answer":    "stub",
		"citations": []any{},
	})
}

// OllamaHealthHandler checks if Ollama service is healthy
func OllamaHealthHandler(w http.ResponseWriter, r *http.Request) {
	client := NewOllamaClient("")
	
	ctx := r.Context()
	if err := client.HealthCheck(ctx); err != nil {
		JSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}
	
	JSON(w, http.StatusOK, map[string]any{
		"status": "healthy",
		"service": "ollama",
	})
}

// OllamaTestHandler tests text generation with Mistral 7B
func OllamaTestHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSON(w, http.StatusBadRequest, map[string]any{
			"error": "Invalid JSON body",
		})
		return
	}
	
	if req.Text == "" {
		req.Text = "Hello, this is a test message for AI processing."
	}
	
	client := NewOllamaClient("")
	ctx := r.Context()
	
	// Test summarization
	summary, err := client.SummarizeEmail(ctx, req.Text)
	if err != nil {
		JSON(w, http.StatusInternalServerError, map[string]any{
			"error": fmt.Sprintf("Summarization failed: %v", err),
		})
		return
	}
	
	// Test classification
	category, err := client.ClassifyEmail(ctx, req.Text)
	if err != nil {
		JSON(w, http.StatusInternalServerError, map[string]any{
			"error": fmt.Sprintf("Classification failed: %v", err),
		})
		return
	}
	
	JSON(w, http.StatusOK, map[string]any{
		"input_text": req.Text,
		"summary":    summary,
		"category":   category,
		"model":      "mistral:7b-instruct",
	})
}
