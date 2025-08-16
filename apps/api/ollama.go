package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaClient provides methods to interact with the Ollama API
type OllamaClient struct {
	baseURL    string
	httpClient *http.Client
}

// GenerateRequest represents a request to generate text
type GenerateRequest struct {
	Model    string   `json:"model"`
	Prompt   string   `json:"prompt"`
	System   string   `json:"system,omitempty"`
	Options  Options  `json:"options,omitempty"`
	Stream   bool     `json:"stream"`
	Format   string   `json:"format,omitempty"`
	Images   []string `json:"images,omitempty"`
	KeepAlive string  `json:"keep_alive,omitempty"`
}

// Options for text generation
type Options struct {
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
	Stop        []string `json:"stop,omitempty"`
	Seed        int      `json:"seed,omitempty"`
}

// GenerateResponse represents the response from text generation
type GenerateResponse struct {
	Model              string    `json:"model"`
	CreatedAt          string    `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	Context            []int     `json:"context,omitempty"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
}

// EmbeddingRequest represents a request to generate embeddings
type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// EmbeddingResponse represents the response from embedding generation
type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(baseURL string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	return &OllamaClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateText generates text using the specified model
func (c *OllamaClient) GenerateText(ctx context.Context, model, prompt, systemPrompt string) (string, error) {
	req := GenerateRequest{
		Model:  model,
		Prompt: prompt,
		System: systemPrompt,
		Stream: false,
		Options: Options{
			Temperature: 0.7,
			TopP:       0.9,
			NumPredict: 500,
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var generateResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&generateResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return generateResp.Response, nil
}

// GenerateEmbedding generates embeddings for the given text
func (c *OllamaClient) GenerateEmbedding(ctx context.Context, model, text string) ([]float64, error) {
	req := EmbeddingRequest{
		Model:  model,
		Prompt: text,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/embeddings", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return embeddingResp.Embedding, nil
}

// HealthCheck checks if the Ollama service is healthy
func (c *OllamaClient) HealthCheck(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// SummarizeEmail generates a summary for an email using Mistral 7B
func (c *OllamaClient) SummarizeEmail(ctx context.Context, emailContent string) (string, error) {
	systemPrompt := `You are an AI assistant that summarizes emails. 
Provide a concise, professional summary in 2-3 sentences. 
Focus on the main points, action items, and key information.`

	return c.GenerateText(ctx, "mistral:7b-instruct", emailContent, systemPrompt)
}

// ClassifyEmail categorizes an email using Mistral 7B
func (c *OllamaClient) ClassifyEmail(ctx context.Context, emailContent string) (string, error) {
	systemPrompt := `You are an AI assistant that classifies emails. 
Analyze the email content and return ONLY ONE of these categories:
- work
- personal  
- important
- spam
- newsletter
- social
- other

Return only the category name, nothing else.`

	return c.GenerateText(ctx, "mistral:7b-instruct", emailContent, systemPrompt)
}
