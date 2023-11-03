package llm

import (
	"context"
)

type ChatModelID string
type EmbeddingModelID string

const (
	ChatModelGPT35Turbo ChatModelID      = "gpt-3.5-turbo"
	ChatModelGPT4       ChatModelID      = "gpt-4"
	EmbeddingModelAda   EmbeddingModelID = "text-embedding-ada-002"
)

type Models struct {
	DefaultChatModel      ChatModel
	DefaultEmbeddingModel EmbeddingModel
	ChatModels            map[ChatModelID]ChatModel
	EmbeddingModels       map[EmbeddingModelID]EmbeddingModel
}

func AllModels(api_key string) Models {
	return Models{
		DefaultChatModel:      NewOpenAIChatModel(ChatModelGPT35Turbo, api_key),
		DefaultEmbeddingModel: NewOpenAIEmbeddingModel(EmbeddingModelAda, api_key),
		ChatModels: map[ChatModelID]ChatModel{
			ChatModelGPT35Turbo: NewOpenAIChatModel(ChatModelGPT35Turbo, api_key),
			ChatModelGPT4:       NewOpenAIChatModel(ChatModelGPT4, api_key),
		},
		EmbeddingModels: map[EmbeddingModelID]EmbeddingModel{
			EmbeddingModelAda: NewOpenAIEmbeddingModel(EmbeddingModelAda, api_key),
		},
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
}

type EmbeddingModel interface {
	Embedding(ctx context.Context, texts []string) ([][]float32, error)
}
