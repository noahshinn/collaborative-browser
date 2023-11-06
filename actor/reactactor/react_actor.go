package reactactor

import (
	"context"
	"webbot/actor"
	"webbot/llm"
)

type ReactActor struct {
	model llm.ChatModel
}

func NewReactActor(model llm.ChatModel) actor.StringActor {
	return &ReactActor{
		model: model,
	}
}

func (ra *ReactActor) Act(ctx context.Context, messages []*llm.Message, functionDefs []*llm.FunctionDef) (string, error) {
	// TODO: implement
	return "", nil
}
