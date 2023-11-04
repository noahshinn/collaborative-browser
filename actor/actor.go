package actor

import (
	"context"
	"encoding/json"
	"fmt"
	"webbot/browser/virtualid"
	"webbot/llm"
	"webbot/trajectory"
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

const systemPromptToActOnBrowser = "You are an AI Assistant that is using a browser. The user can see the same browser as you. Your job is to take actions on the browser as requested by the user."

var permissibleFunctions = []*llm.FunctionDef{
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
					Description: "The reason that the task is complete",
				},
			},
			Required: []string{"reason"},
		},
	},
	{
		Name: "task_not_possible",
		Parameters: llm.Parameters{
			Type: "object",
			Properties: map[string]llm.Property{
				"reason": {
					Type:        "string",
					Description: "The reason that it is not possible to complete the task",
				},
			},
			Required: []string{"reason"},
		},
	},
}

func permissibleFunctionMap() map[string]*llm.FunctionDef {
	m := make(map[string]*llm.FunctionDef)
	for _, functionDef := range permissibleFunctions {
		m[functionDef.Name] = functionDef
	}
	return m
}

const maxTokenContextWindowMarginProportion float32 = 0.1

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
	approxNumTokens := llm.ApproxNumTokensInMessages(messages) + llm.ApproxNumTokensInFunctionDefs(permissibleFunctions)
	if approxNumTokens > int(float32(a.ChatModel.ContextLength())*(1-maxTokenContextWindowMarginProportion)) {
		return trajectory.NewErrorMaxContextLengthExceeded(a.ChatModel.ContextLength(), approxNumTokens), nil
	}
	if res, err := a.ChatModel.Message(ctx, messages, &llm.MessageOptions{
		Temperature: 0.0,
		Functions:   permissibleFunctions,
	}); err != nil {
		return nil, err
	} else if res.FunctionCall == nil {
		return trajectory.NewAgentMessage(res.Content), nil
	} else if _, ok := permissibleFunctionMap()[res.FunctionCall.Name]; !ok {
		return nil, fmt.Errorf("unsupported action was attempted: %s", res.FunctionCall.Name)
	} else if nextAction, err := parseNextAction(res.FunctionCall.Name, res.FunctionCall.Arguments); err != nil {
		return nil, err
	} else {
		return nextAction, nil
	}
}

func parseNextAction(name string, arguments string) (trajectory.TrajectoryItem, error) {
	var args map[string]any
	functions := permissibleFunctionMap()
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return nil, fmt.Errorf("error unmarshaling function call arguments: %w", err)
	}
	if _, ok := functions[name]; !ok {
		return nil, fmt.Errorf("unsupported action was attempted: %s", name)
	}
	for _, required := range functions[name].Parameters.Required {
		if _, ok := args[required]; !ok {
			return nil, fmt.Errorf("required argument %s was not supplied", required)
		}
	}
	for argName := range args {
		if _, ok := functions[name].Parameters.Properties[argName]; !ok {
			return nil, fmt.Errorf("unsupported argument %s was supplied", argName)
		}
	}
	switch name {
	case "click":
		return trajectory.NewBrowserClickAction(virtualid.VirtualID(args["id"].(string))), nil
	case "send_keys":
		return trajectory.NewBrowserSendKeysAction(virtualid.VirtualID(args["id"].(string)), args["text"].(string)), nil
	case "navigate":
		return trajectory.NewBrowserNavigateAction(args["url"].(string)), nil
	case "message":
		return trajectory.NewAgentMessage(args["text"].(string)), nil
	case "task_complete":
		return trajectory.NewBrowserTaskCompleteAction(args["reason"].(string)), nil
	case "task_not_possible":
		return trajectory.NewBrowserTaskNotPossibleAction(args["reason"].(string)), nil
	default:
		return nil, fmt.Errorf("unsupported action was attempted: %s", name)
	}
}
