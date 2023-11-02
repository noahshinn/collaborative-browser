package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const OPENAI_API_URL = "https://api.openai.com/v1"

type OpenAIModel struct {
	modelID ChatModelID
	apiKey  string
}

type OpenAIEmbeddingModel struct {
	modelID EmbeddingModelID
	apiKey  string
}

func NewOpenAIChatModel(modelID ChatModelID, apiKey string) ChatModel {
	return &OpenAIModel{modelID: modelID, apiKey: apiKey}
}

func NewOpenAIEmbeddingModel(embeddingModelID EmbeddingModelID, apiKey string) EmbeddingModel {
	return &OpenAIEmbeddingModel{modelID: embeddingModelID, apiKey: apiKey}
}

func (m *OpenAIModel) MessageStream(ctx context.Context, messages []*Message, options *MessageOptions) (chan StreamEvent, error) {
	// TODO: implement
	panic("not implemented")
}

func (m *OpenAIModel) Message(ctx context.Context, messages []*Message, options *MessageOptions) (*Message, error) {
	args := m.buildArgs(messages, options)
	if response, err := apiRequest(ctx, m.apiKey, "/chat/completions", args); err != nil {
		return nil, err
	} else {
		return parseResponse(response)
	}
}

func (m *OpenAIModel) buildArgs(messages []*Message, options *MessageOptions) map[string]any {
	jsonMessages := []map[string]string{}
	for _, message := range messages {
		jsonMessage := map[string]string{
			"role":    string(message.Role),
			"content": message.Content,
		}
		if message.Name != "" {
			jsonMessage["name"] = message.Name
		}
		jsonMessages = append(jsonMessages, jsonMessage)
	}
	args := map[string]any{
		"model":       m.modelID,
		"messages":    jsonMessages,
		"temperature": options.Temperature,
	}
	if options.MaxTokens > 0 {
		args["max_tokens"] = options.MaxTokens
	}
	if len(options.StopSequences) > 0 {
		args["stop"] = options.StopSequences
	}
	if len(options.Functions) > 0 {
		args["functions"] = options.Functions
	}
	if options.FunctionCall != "" {
		if options.FunctionCall == FunctionCallNone || options.FunctionCall == FunctionCallAuto {
			args["function_call"] = options.FunctionCall
		} else {
			args["function_call"] = fmt.Sprintf("{\"name\":\\ \"%s\"}", options.FunctionCall)
		}
	}
	return args
}

func (m *OpenAIModel) Action(ctx context.Context, messages []*Message, options *MessageOptions) (*Action, error) {
	args := m.buildArgs(messages, options)
	if response, err := apiRequest(ctx, m.apiKey, "/chat/completions", args); err != nil {
		return nil, err
	} else if message, err := parseResponse(response); err != nil {
		return nil, err
	} else if message.FunctionCall == nil {
		return nil, &Error{Message: "invalid response, no function call"}
	} else {
		return &Action{
			Name:      message.FunctionCall.Name,
			Arguments: message.FunctionCall.Arguments,
		}, nil
	}
}

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func parseResponse(response map[string]any) (*Message, error) {
	if choices, ok := response["choices"].([]any); !ok {
		return nil, &Error{Message: "invalid response, no choices"}
	} else if len(choices) != 1 {
		return nil, &Error{Message: "invalid response, expected 1 choice"}
	} else if choice, ok := choices[0].(map[string]any); !ok {
		return nil, &Error{Message: "invalid response, choice is not a map"}
	} else if message, ok := choice["message"].(map[string]any); !ok {
		return nil, &Error{Message: "invalid response, message is not a map"}
	} else if content, ok := message["content"].(string); ok {
		return &Message{
			Role:    MessageRole(message["role"].(string)),
			Content: content,
		}, nil
	} else if functionCall, ok := choice["function_call"].(map[string]any); ok {
		if name, ok := functionCall["name"].(string); !ok {
			return nil, &Error{Message: "invalid response, function call has no name"}
		} else if functionCallArgs, ok := functionCall["args"].(string); !ok {
			return nil, &Error{Message: "invalid response, function call args is not a map"}
		} else {
			return &Message{
				Role: MessageRoleFunction,
				Name: name,
				FunctionCall: &FunctionCall{
					Name:      name,
					Arguments: functionCallArgs,
				},
			}, nil
		}
	}
	return nil, &Error{Message: "invalid response, no content or function call"}
}

func apiRequest(ctx context.Context, apiKey string, endpoint string, args map[string]any) (map[string]any, error) {
	if encoded, err := json.Marshal(args); err != nil {
		return nil, err
	} else if request, err := http.NewRequestWithContext(ctx, "POST", OPENAI_API_URL+endpoint, bytes.NewBuffer(encoded)); err != nil {
		return nil, err
	} else {
		request.Header.Set("Content-Type", "application/json; charset=utf-8")
		request.Header.Set("Authorization", "Bearer "+apiKey)
		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			return nil, err
		} else if responseBody, err := io.ReadAll(response.Body); err != nil {
			return nil, err
		} else {
			result := map[string]any{}
			if err := json.Unmarshal(responseBody, &result); err != nil {
				return nil, err
			}
			if err, ok := result["error"].(map[string]any); ok {
				response := Error{Message: "OpenAI error"}
				if value, ok := err["code"].(string); ok {
					response.Code = value
				}
				if value, ok := err["message"].(string); ok {
					response.Message = value
				}
				return nil, &response
			}
			return result, nil
		}
	}
}

func (m *OpenAIEmbeddingModel) Embedding(ctx context.Context, text []string) ([]float32, error) {
	// TODO: implement
	return nil, nil
}
