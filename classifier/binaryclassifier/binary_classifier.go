package binaryclassifier

import (
	"context"
	"encoding/json"
	"fmt"
	"webbot/llm"
)

type BinaryClassifier interface {
	Classify(ctx context.Context, text string) (bool, error)
}

func LLMBinaryClassification(ctx context.Context, model llm.ChatModel, systemPromptText string, classificationInstruction string, text string, withReasoning bool) (bool, error) {
	messages := []*llm.Message{
		{
			Role:    llm.MessageRoleSystem,
			Content: systemPromptText,
		},
		{
			Role:    llm.MessageRoleUser,
			Content: fmt.Sprintf("# Text to classify:\n%s\n\n# Classification instruction:\n%s", text, classificationInstruction),
		},
	}
	var classificationFunctionDef *llm.FunctionDef
	if withReasoning {
		classificationFunctionDef = &llm.FunctionDef{
			Name: "classify",
			Parameters: llm.Parameters{
				Type: "object",
				Properties: map[string]llm.Property{
					"classification": {
						Type:        "boolean",
						Description: "The classification decision",
					},
				},
				Required: []string{"classification"},
			},
		}
	} else {
		classificationFunctionDef = &llm.FunctionDef{
			Name: "classify",
			Parameters: llm.Parameters{
				Type: "object",
				Properties: map[string]llm.Property{
					"reasoning": {
						Type:        "string",
						Description: "The reasoning for the classification decision",
					},
					"classification": {
						Type:        "boolean",
						Description: "The classification decision",
					},
				},
				Required: []string{"reasoning", "classification"},
			},
		}
	}
	if res, err := model.Message(ctx, messages, &llm.MessageOptions{
		Functions:    []*llm.FunctionDef{classificationFunctionDef},
		Temperature:  0.0,
		FunctionCall: "classify",
	}); err != nil {
		return false, fmt.Errorf("failed to classify text: %w", err)
	} else if res.FunctionCall == nil {
		return false, fmt.Errorf("failed to classify text: no function call returned")
	} else if res.FunctionCall.Name != "classify" {
		return false, fmt.Errorf("failed to classify text: unexpected function call name: %s", res.FunctionCall.Name)
	} else if res.FunctionCall.Arguments == "" {
		return false, fmt.Errorf("failed to classify text: no function call arguments returned")
	} else {
		var args map[string]any
		if err := json.Unmarshal([]byte(res.FunctionCall.Arguments), &args); err != nil {
			return false, fmt.Errorf("failed to classify text: error unmarshaling function call arguments: %w", err)
		}
		if classification, ok := args["classification"]; !ok {
			return false, fmt.Errorf("failed to classify text: no classification returned")
		} else if classificationBool, ok := classification.(bool); !ok {
			return false, fmt.Errorf("failed to classify text: classification was not a boolean")
		} else {
			return classificationBool, nil
		}
	}
}
