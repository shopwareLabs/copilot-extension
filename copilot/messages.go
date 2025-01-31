package copilot

import "github.com/invopop/jsonschema"

type ChatRequest struct {
	Messages []ChatMessage `json:"messages"`
}

type ChatMessage struct {
	Role          string              `json:"role"`
	Content       string              `json:"content"`
	Confirmations []*ChatConfirmation `json:"copilot_confirmations"`
	ToolCalls     []*ToolCall         `json:"tool_calls"`
}

type ToolCall struct {
	Function *ChatMessageFunctionCall `json:"function"`
}

type ChatMessageFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ChatConfirmation struct {
	State        string            `json:"state"`
	Confirmation *ConfirmationData `json:"confirmation"`
}

type ConfirmationData struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type Model string

const (
	ModelGPT35      Model = "gpt-3.5-turbo"
	ModelGPT4       Model = "gpt-4"
	ModelGPT4O      Model = "gpt-40"
	ModelEmbeddings Model = "text-embedding-ada-002"
)

type ChatCompletionsRequest struct {
	Messages []ChatMessage  `json:"messages"`
	Model    Model          `json:"model"`
	Tools    []FunctionTool `json:"tools"`
	Stream   bool           `json:"stream"`
}

type FunctionTool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Parameters  *jsonschema.Schema `json:"parameters"`
}

type ChatCompletionsResponse struct {
	Choices []struct {
		FinishReason         string `json:"finish_reason"`
		Index                int    `json:"index"`
		ContentFilterOffsets struct {
			CheckOffset int `json:"check_offset"`
			StartOffset int `json:"start_offset"`
			EndOffset   int `json:"end_offset"`
		} `json:"content_filter_offsets"`
		ContentFilterResults struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
			Hate struct {
				Filtered bool   `json:"filtered"`
				Severity string `json:"severity"`
			} `json:"hate"`
			SelfHarm struct {
				Filtered bool   `json:"filtered"`
				Severity string `json:"severity"`
			} `json:"self_harm"`
			Sexual struct {
				Filtered bool   `json:"filtered"`
				Severity string `json:"severity"`
			} `json:"sexual"`
			Violence struct {
				Filtered bool   `json:"filtered"`
				Severity string `json:"severity"`
			} `json:"violence"`
		} `json:"content_filter_results"`
		Delta struct {
			Content   string `json:"content"`
			ToolCalls []struct {
				Function *ChatMessageFunctionCall `json:"function"`
				ID       string                   `json:"id"`
				Index    int                      `json:"index"`
				Type     string                   `json:"type"`
			} `json:"tool_calls"`
		} `json:"delta"`
	} `json:"choices"`
	Created           int    `json:"created"`
	ID                string `json:"id"`
	Model             string `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
}
