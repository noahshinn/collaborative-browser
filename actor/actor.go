package actor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"webbot/browser/virtualid"
	"webbot/llm"
	"webbot/runner/trajectory"
	"webbot/utils/slicesx"
)

type Actor interface {
	NextAction(ctx context.Context, state string) (trajectory.TrajectoryItem, error)
}

type LLMActor struct {
	ChatModel llm.ChatModel
}

func NewLLMActor(chatModel llm.ChatModel) Actor {
	return &LLMActor{ChatModel: chatModel}
}

const systemPromptToActOnBrowser = "Take an action on the browser."

func (a *LLMActor) NextAction(ctx context.Context, state string) (trajectory.TrajectoryItem, error) {
	messages := []*llm.Message{
		{
			Role:    llm.MessageRoleSystem,
			Content: systemPromptToActOnBrowser,
		},
		{
			Role:    llm.MessageRoleUser,
			Content: state,
		},
	}
	functions := []llm.FunctionDef{
		{
			Name: "click",
			Parameters: llm.Parameters{
				Type: "object",
				Properties: map[string]llm.Property{
					"id": {
						Type:        "string",
						Description: "The id of the element to click",
					},
				},
				Required: []string{"id"},
			},
		},
		{
			Name: "send_keys",
			Parameters: llm.Parameters{
				Type: "object",
				Properties: map[string]llm.Property{
					"id": {
						Type:        "string",
						Description: "The id of the element to send keys to",
					},
					"text": {
						Type:        "string",
						Description: "The text to send to the element",
					},
				},
				Required: []string{"id", "text"},
			},
		},
		{
			Name: "navigate",
			Parameters: llm.Parameters{
				Type: "object",
				Properties: map[string]llm.Property{
					"url": {
						Type:        "string",
						Description: "The url to navigate the browser to",
					},
				},
				Required: []string{"url"},
			},
		},
		{
			Name: "message",
			Parameters: llm.Parameters{
				Type: "object",
				Properties: map[string]llm.Property{
					"text": {
						Type:        "string",
						Description: "The text for a message to send to the user",
					},
				},
				Required: []string{"text"},
			},
		},
		{
			Name: "task_complete",
			Parameters: llm.Parameters{
				Type: "object",
				Properties: map[string]llm.Property{
					"reason": {
						Type:        "string",
						Description: "The reason that the task is complete or cannot be completed",
					},
				},
				Required: []string{"reason"},
			},
		},
	}
	permissibleActionNames := slicesx.Map(functions, func(function llm.FunctionDef) string { return function.Name })
	if res, err := a.ChatModel.Message(ctx, messages, &llm.MessageOptions{
		Temperature: 0.0,
		Functions:   functions,
	}); err != nil {
		return nil, err
	} else if res.FunctionCall == nil {
		return nil, errors.New("no function call")
	} else if res.FunctionCall.Name == "" || !slicesx.Contains(permissibleActionNames, res.FunctionCall.Name) {
		return nil, fmt.Errorf("unsupported action was attempted: %s", res.FunctionCall.Name)
	} else {
		var args map[string]any
		if err := json.Unmarshal([]byte(res.FunctionCall.Arguments), &args); err != nil {
			return nil, fmt.Errorf("error unmarshaling function call arguments: %w", err)
		}
		switch res.FunctionCall.Name {
		case "click":
			if id, ok := args["id"].(string); !ok {
				return nil, errors.New("\"click\" action was taken but no id was supplied")
			} else {
				return trajectory.NewBrowserClickAction(virtualid.VirtualID(id)), nil
			}
		case "send_keys":
			if id, ok := args["id"].(string); !ok {
				return nil, errors.New("\"send_keys\" action was taken but no id was supplied")
			} else if text, ok := args["text"].(string); !ok || text == "" {
				return nil, errors.New("\"send_keys\" action was taken but no text was supplied")
			} else {
				return trajectory.NewBrowserSendKeysAction(virtualid.VirtualID(id), text), nil
			}
		case "navigate":
			if url, ok := args["url"].(string); !ok {
				return nil, errors.New("\"navigate\" action was taken but no url was supplied")
			} else {
				return trajectory.NewBrowserNavigateAction(url), nil
			}
		case "message":
			if text, ok := args["text"].(string); !ok || text == "" {
				return nil, errors.New("\"message\" action was taken but no text was supplied")
			} else {
				return trajectory.NewAgentMessage(text), nil
			}
		default:
			return nil, fmt.Errorf("unsupported action was attempted: %s", res.FunctionCall.Name)
		}
	}
}
