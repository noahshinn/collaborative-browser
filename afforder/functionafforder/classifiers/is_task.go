package classifiers

import (
	"context"
	"webbot/llm"
)

type IsTask struct {
	systemPrompt        string
	instruction         string
	classificationModel llm.ChatModel
	withReasoning       bool
}

// TODO: implement
const systemPrompt = ""

const instruction = ""

func NewIsTaskClassifier(model llm.ChatModel, withReasoning bool) BinaryClassifier {
	return &IsTask{
		systemPrompt:        systemPrompt,
		instruction:         instruction,
		classificationModel: model,
		withReasoning:       withReasoning,
	}
}

func (bc *IsTask) Classify(ctx context.Context, text string) (bool, error) {
	return LLMBinaryClassification(ctx, bc.classificationModel, bc.systemPrompt, bc.instruction, text, bc.withReasoning)
}
