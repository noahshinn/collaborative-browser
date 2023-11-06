package reflexionactor

import (
	"context"
	"fmt"
	"webbot/actor"
	"webbot/actor/llmactor"
	"webbot/llm"
)

type ReflexionActor struct {
	maxNumIterations int
	model            llm.ChatModel
	llmActor         actor.StringActor
}

type Options struct {
	MaxNumIterations int
	LLMActor         actor.StringActor
}

func NewReflexionActor(model llm.ChatModel, options *Options) actor.StringActor {
	maxNumIterations := 1
	llmActor := llmactor.NewLLMActor(model)
	if options != nil {
		if options.MaxNumIterations > 1 {
			maxNumIterations = options.MaxNumIterations
		}
		if options.LLMActor != nil {
			llmActor = options.LLMActor
		}
	}
	return &ReflexionActor{
		maxNumIterations: maxNumIterations,
		model:            model,
		llmActor:         llmActor,
	}
}

func (a *ReflexionActor) Act(ctx context.Context, messages []*llm.Message, functionDefs []*llm.FunctionDef) (string, error) {
	nextAction, err := a.llmActor.Act(ctx, messages, functionDefs)
	if err != nil {
		return "", fmt.Errorf("error acting with llm actor: %w", err)
	}
	// TODO: do reflection in a loop or once
	return nextAction, nil
}
