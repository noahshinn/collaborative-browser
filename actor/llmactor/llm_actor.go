package llmactor

import (
	"context"
	"webbot/actor"
	"webbot/llm"
)

type LLMActor struct {
	model llm.ChatModel
}

func NewLLMActor(model llm.ChatModel) actor.StringActor {
	return &LLMActor{
		model: model,
	}
}

func (a *LLMActor) Act(ctx context.Context, messages []*llm.Message, functionDefs []*llm.FunctionDef) (string, error) {
	_, err := a.model.Message(ctx, messages, &llm.MessageOptions{
		Temperature: 0.0,
		Functions:   functionDefs,
	})
	if err != nil {
		return "", err
	}
	// TODO: parse
	return "", nil
}
