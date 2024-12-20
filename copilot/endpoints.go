package copilot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

func ChatCompletions(ctx context.Context, client *retryablehttp.Client, integrationID, apiKey string, req *ChatCompletionsRequest) (*ChatCompletionsResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.githubcopilot.com/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	if integrationID != "" {
		httpReq.Header.Set("Copilot-Integration-Id", integrationID)
	}

	resp, err := client.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var chatRes *ChatCompletionsResponse
	err = json.NewDecoder(resp.Body).Decode(&chatRes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return chatRes, nil
}

func Embeddings(ctx context.Context, client *retryablehttp.Client, integrationID string, token string, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.githubcopilot.com/embeddings", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	if integrationID != "" {
		httpReq.Header.Set("Copilot-Integration-Id", integrationID)
	}

	resp, err := client.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var embeddingsResponse *EmbeddingsResponse
	err = json.NewDecoder(resp.Body).Decode(&embeddingsResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode request body")
	}

	return embeddingsResponse, nil
}
