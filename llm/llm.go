package llm

import (
	"context"
	"fmt"
)

type ChatModelID string
type EmbeddingModelID string

const (
	ChatModelGPT35Turbo ChatModelID      = "gpt-3.5-turbo"
	ChatModelGPT4       ChatModelID      = "gpt-4"
	EmbeddingModelAda   EmbeddingModelID = "text-embedding-ada-002"
)

type Models struct {
	ChatModels      map[ChatModelID]ChatModel
	EmbeddingModels map[EmbeddingModelID]EmbeddingModel
}

var AllModels = Models{
	ChatModels: map[ChatModelID]ChatModel{
		ChatModelGPT35Turbo: NewOpenAIChatModel(ChatModelGPT35Turbo, "api-key"),
		ChatModelGPT4:       NewOpenAIChatModel(ChatModelGPT4, "api-key"),
	},
	EmbeddingModels: map[EmbeddingModelID]EmbeddingModel{
		EmbeddingModelAda: NewOpenAIEmbeddingModel(EmbeddingModelAda, "api-key"),
	},
}

func (m *Models) ChatModel(id string) (ChatModel, error) {
	if model, ok := m.ChatModels[ChatModelID(id)]; !ok {
		return nil, fmt.Errorf("chat model %s not found", id)
	} else {
		return model, nil
	}
}

func (m *Models) EmbeddingModel(id string) (EmbeddingModel, error) {
	if model, ok := m.EmbeddingModels[EmbeddingModelID(id)]; !ok {
		return nil, fmt.Errorf("embedding model %s not found", id)
	} else {
		return model, nil
	}
}

type FunctionDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  Parameters `json:"parameters"`
}

type Parameters struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type FunctionCall struct {
	Name      string
	Arguments string
}

type Message struct {
	Role         MessageRole   `json:"role"`
	Content      string        `json:"content"`
	Name         string        `json:"name"`
	FunctionCall *FunctionCall `json:"function_call"`
}

type MessageRole string

const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleFunction  MessageRole = "function"
)

type Action struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type StreamEvent struct {
	Text  string
	Error error
}

type MessageOptions struct {
	Temperature   float32  `json:"temperature"`
	MaxTokens     int      `json:"max_tokens"`
	StopSequences []string `json:"stop_sequences"`

	// for OpenAI models
	Functions    []FunctionDef `json:"functions"`
	FunctionCall string        `json:"function_call"`
}

const FunctionCallNone = "none"
const FunctionCallAuto = "auto"

type ChatModel interface {
	MessageStream(ctx context.Context, messages []*Message, options *MessageOptions) (chan StreamEvent, error)
	Message(ctx context.Context, messages []*Message, options *MessageOptions) (*Message, error)
	Action(ctx context.Context, messages []*Message, options *MessageOptions) (*Action, error)
}

type EmbeddingModel interface {
	Embedding(ctx context.Context, text []string) ([]float32, error)
}
