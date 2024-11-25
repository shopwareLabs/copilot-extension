package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/invopop/jsonschema"
	"github.com/shopwarelabs/copilot-extension/copilot"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type Tools []copilot.FunctionTool

func (t Tools) RemoveTool(name []string) Tools {
	scopedTools := make(Tools, 0)

	for _, tool := range t {
		if !slices.Contains(name, tool.Function.Name) {
			scopedTools = append(scopedTools, tool)
		}
	}

	return scopedTools
}

var tools Tools
var loadShopwareVersions sync.RWMutex
var shopwareVersions string

func init() {
	releaseNotes := orderedmap.New[string, *jsonschema.Schema]()
	releaseNotes.Set("version", &jsonschema.Schema{
		Type:        "string",
		Description: "The Shopware version to get the release notes for",
	})

	storePlugin := orderedmap.New[string, *jsonschema.Schema]()
	storePlugin.Set("name", &jsonschema.Schema{
		Type:        "array",
		Description: "The name of the extension/plugin/app in the Shopware Store",
		Items: &jsonschema.Schema{
			Type: "string",
		},
	})

	tools = []copilot.FunctionTool{
		{
			Type: "function",
			Function: copilot.Function{
				Name:        "get_shopware_versions",
				Description: "Get all available Shopware versions",
			},
		},
		{
			Type: "function",
			Function: copilot.Function{
				Name:        "get_release_notes",
				Description: "Get the release notes or changelog for a specific Shopware version",
				Parameters: &jsonschema.Schema{
					Type:       "object",
					Properties: releaseNotes,
					Required:   []string{"version"},
				},
			},
		},
		{
			Type: "function",
			Function: copilot.Function{
				Name:        "get_store_extension",
				Description: "Get name, description and changelog of multiple extension/plugin/app in the Shopware Store",
				Parameters: &jsonschema.Schema{
					Type:       "object",
					Properties: storePlugin,
					Required:   []string{"name"},
				},
			},
		},
	}
}

func handleFunction(ctx context.Context, function *copilot.ChatMessageFunctionCall) (*copilot.ChatMessage, error) {
	switch function.Name {
	case "get_shopware_versions":
		return getShopwareVersions(ctx)
	case "get_release_notes":
		return getReleaseNotes(ctx, function.Arguments)
	case "get_store_extension":
		return getStoreExtension(ctx, function.Arguments)
	default:
		return nil, fmt.Errorf("unknown function: %s", function.Name)
	}
}

type GithubRelease struct {
	TagName     string `json:"tag_name"`
	PublishedAt string `json:"published_at"`
}

func getShopwareVersions(ctx context.Context) (*copilot.ChatMessage, error) {
	loadShopwareVersions.RLock()

	if shopwareVersions != "" {
		loadShopwareVersions.RUnlock()
		return &copilot.ChatMessage{
			Role:    "system",
			Content: shopwareVersions,
		}, nil
	}

	loadShopwareVersions.RUnlock()

	loadShopwareVersions.Lock()
	defer loadShopwareVersions.Unlock()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/shopware/shopware/releases?per_page=100", nil)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var releases []GithubRelease

	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	for _, release := range releases {
		shopwareVersions += fmt.Sprintf("%s released at: %s\n", release.TagName, release.PublishedAt)
	}

	return &copilot.ChatMessage{
		Role:    "system",
		Content: shopwareVersions,
	}, nil
}

func getReleaseNotes(ctx context.Context, arguments string) (*copilot.ChatMessage, error) {
	var parameters struct {
		Version string `json:"version"`
	}

	if err := json.Unmarshal([]byte(arguments), &parameters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	normalizedVersion := strings.TrimPrefix(parameters.Version, "v")

	shortVersion := normalizedVersion[0:3]

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://raw.githubusercontent.com/shopware/release-notes/refs/heads/main/src/%s/%s.md", shortVersion, normalizedVersion), nil)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &copilot.ChatMessage{
			Role:    "system",
			Content: "",
		}, nil
	}

	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)

	if err != nil {
		return &copilot.ChatMessage{
			Role:    "system",
			Content: "",
		}, nil
	}

	return &copilot.ChatMessage{
		Role:    "system",
		Content: string(content),
	}, nil
}

func getStoreExtension(ctx context.Context, arguments string) (*copilot.ChatMessage, error) {
	var parameters struct {
		Name []string `json:"name"`
	}

	if err := json.Unmarshal([]byte(arguments), &parameters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	u, _ := url.Parse("https://api.shopware.com/pluginStore/pluginsByName?shopwareVersion=6.6.8.2&locale=en-GB")

	query := u.Query()

	for _, name := range parameters.Name {
		query.Add("technicalNames[]", name)
	}

	u.RawQuery = query.Encode()

	log.Infof("URL: %s", u.String())

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var plugins []struct {
		Label       string `json:"label"`
		Description string `json:"description"`
		Version     string `json:"version"`
		Changelog   []struct {
			Version string `json:"version"`
			Text    string `json:"text"`
		} `json:"changelog"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&plugins); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	var content string

	for _, plugin := range plugins {
		content += fmt.Sprintf("# %s\n\n", plugin.Label)

		if plugin.Description != "" {
			content += fmt.Sprintf("## Description\n\n%s\n\n", plugin.Description)
		}

		if len(plugin.Changelog) > 0 {
			content += "## Changelog\n\n"

			for _, changelog := range plugin.Changelog {
				content += fmt.Sprintf("### %s\n\n%s\n\n", changelog.Version, changelog.Text)
			}
		}
	}

	return &copilot.ChatMessage{
		Role:    "system",
		Content: string(content),
	}, nil
}
