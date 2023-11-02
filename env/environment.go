package env

import "webbot/llm"

type Environment struct {
	DefaultMessageModel   llm.ChatModel
	DefaultEmbeddingModel llm.EmbeddingModel
}

func NewEnvironment() *Environment {
	return &Environment{
		DefaultMessageModel:   llm.AllModels.ChatModels[llm.ChatModelGPT35Turbo],
		DefaultEmbeddingModel: llm.AllModels.EmbeddingModels[llm.EmbeddingModelAda],
	}
}
