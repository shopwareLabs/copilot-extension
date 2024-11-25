package agent

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/philippgille/chromem-go"
	"github.com/shopwarelabs/copilot-extension/copilot"
)

var fileRegexp = regexp.MustCompile(`(?m)^(.*\.\w+)_\d+$`)

// Service provides and endpoint for this agent to perform chat completions
type Service struct {
	pubKey     *ecdsa.PublicKey
	collection *chromem.Collection
}

func NewService(pubKey *ecdsa.PublicKey, collection *chromem.Collection) *Service {
	return &Service{
		pubKey:     pubKey,
		collection: collection,
	}
}

func (s *Service) ChatCompletion(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Infof("failed to read request body: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Make sure the payload matches the signature. In this way, you can be sure
	// that an incoming request comes from github

	isValid, err := validPayload(body, r.Header.Get("Github-Public-Key-Signature"), s.pubKey)
	if err != nil {
		log.Infof("failed to validate payload signature: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !isValid {
		http.Error(w, "invalid payload signature", http.StatusUnauthorized)
		return
	}

	apiToken := r.Header.Get("X-GitHub-Token")
	integrationID := r.Header.Get("Copilot-Integration-Id")

	var req *copilot.ChatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Infof("failed to unmarshal request: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.generateCompletion(r.Context(), integrationID, apiToken, req, NewSSEWriter(w)); err != nil {
		log.Infof("failed to execute agent: %v\n", err)
	}
}

func (s *Service) generateCompletion(ctx context.Context, integrationID, apiToken string, req *copilot.ChatRequest, w *sseWriter) error {
	var messages []copilot.ChatMessage
	copilotReferences := []sseReference{}

	messages = append(messages, req.Messages...)

	// Create embeddings from user messages
	for i := len(req.Messages) - 1; i >= 0; i++ {
		msg := req.Messages[i]
		if msg.Role != "user" {
			continue
		}

		// Filter empty messages
		if msg.Content == "" {
			continue
		}

		startTime := time.Now()

		embeddings, err := copilot.Embeddings(ctx, retryablehttp.NewClient(), integrationID, apiToken, &copilot.EmbeddingsRequest{
			Model: copilot.ModelEmbeddings,
			Input: []string{msg.Content},
		})

		if err != nil {
			return fmt.Errorf("failed to get embeddings: %w", err)
		}

		log.Infof("Embeddings took %s", time.Since(startTime))

		startTime = time.Now()

		res, err := s.collection.QueryEmbedding(ctx, embeddings.Data[0].Embedding, 5, nil, nil)

		if err != nil {
			return fmt.Errorf("failed to query collection: %w", err)
		}

		log.Infof("Query took %s", time.Since(startTime))

		contextMessage := ""

		for _, doc := range res {
			link := "unknown"
			fileName := doc.ID

			if strings.HasPrefix(doc.ID, "data/docs/") {
				fileName = fileRegexp.FindStringSubmatch(strings.TrimPrefix(doc.ID, "data/docs/"))[1]

				link = fmt.Sprintf("https://github.com/shopware/docs/blob/main/%s", fileName)
			} else if strings.HasPrefix(doc.ID, "data/src/") {
				fileName = fileRegexp.FindStringSubmatch(strings.TrimPrefix(doc.ID, "data/"))[1]

				link = fmt.Sprintf("https://github.com/shopware/shopware/blob/trunk/%s", fileName)
			}

			copilotReferences = append(copilotReferences, sseReference{
				Type: "document",
				ID:   doc.ID,
				Metadata: sseReferenceMetadata{
					DisplayName: fileName,
					DisplayIcon: "icon",
					DisplayURL:  link,
				},
			})

			contextMessage += doc.Content + "\n"
		}

		messages = append(messages, copilot.ChatMessage{
			Role: "system",
			Content: "You are a specialized technical chatbot for Shopware 6 development. Your primary goal is to assist developers with precise and accurate technical information about Shopware 6. Always provide detailed, developer-focused responses that cover both theoretical concepts and practical implementation. When asked, generate relevant code examples and explain them thoroughly, including best practices for Shopware 6 development. Your knowledge is based on the provided Shopware 6 documentation and code examples. If you're unsure about something, admit it and suggest where the user might find more information. Respond in a clear, concise, and technical manner suitable for developers. Use proper formatting for code snippets and technical terms. When explaining concepts, break them down into easily understandable parts. If providing step-by-step instructions, number them clearly. Always strive for accuracy and completeness in your responses. If a question is ambiguous, ask for clarification to ensure you provide the most relevant information.\n" +
				"Context: " + contextMessage + "\nWhen calling get_store_extension pass all app/plugin/extension names",
		})

		break
	}

	usedTools := []string{}

	for {
		chatReq := &copilot.ChatCompletionsRequest{
			Model:    copilot.ModelGPT4,
			Messages: messages,
			Tools:    tools.RemoveTool(usedTools),
		}

		res, err := copilot.ChatCompletions(ctx, retryablehttp.NewClient(), integrationID, apiToken, chatReq)
		if err != nil {
			return fmt.Errorf("failed to get chat completions stream: %w", err)
		}

		function := getFunctionCall(res)

		if function != nil {
			usedTools = append(usedTools, function.Name)
			log.Debugf("Function CALL: %s", function.Name)

			msg, err := handleFunction(ctx, function)

			if err != nil {
				w.writeEvent("copilot_errors")
				w.writeData([]sseError{{Type: "function", Code: "failed", Message: err.Error(), Identifier: function.Name}})
				w.writeDone()

				return fmt.Errorf("failed to handle function: %w", err)
			}

			messages = append(messages, *msg)

			log.Infof("Responded function call")

			continue
		} else {
			log.Infof("Responded normal message")
			choices := make([]sseResponseChoice, len(res.Choices))
			for i, choice := range res.Choices {
				choices[i] = sseResponseChoice{
					Index: choice.Index,
					Delta: sseResponseMessage{
						Role:    choice.Message.Role,
						Content: choice.Message.Content,
					},
				}
			}

			// w.writeEvent("copilot_references")
			// w.writeData(copilotReferences)

			// w.writeDone()

			w.writeData(sseResponse{
				Choices: choices,
			})

			w.writeDone()

			break
		}
	}

	return nil
}

// asn1Signature is a struct for ASN.1 serializing/parsing signatures.
type asn1Signature struct {
	R *big.Int
	S *big.Int
}

func validPayload(data []byte, sig string, publicKey *ecdsa.PublicKey) (bool, error) {
	asnSig, err := base64.StdEncoding.DecodeString(sig)
	parsedSig := asn1Signature{}
	if err != nil {
		return false, err
	}
	rest, err := asn1.Unmarshal(asnSig, &parsedSig)
	if err != nil || len(rest) != 0 {
		return false, err
	}

	// Verify the SHA256 encoded payload against the signature with GitHub's Key
	digest := sha256.Sum256(data)
	return ecdsa.Verify(publicKey, digest[:], parsedSig.R, parsedSig.S), nil
}

func getFunctionCall(res *copilot.ChatCompletionsResponse) *copilot.ChatMessageFunctionCall {
	if len(res.Choices) == 0 {
		return nil
	}

	if len(res.Choices[0].Message.ToolCalls) == 0 {
		return nil
	}

	funcCall := res.Choices[0].Message.ToolCalls[0].Function
	if funcCall == nil {
		return nil
	}
	return funcCall

}
