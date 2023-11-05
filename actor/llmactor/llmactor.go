package llmactor

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"webbot/actor"
	"webbot/browser"
	"webbot/browser/language"
	"webbot/browser/virtualid"
	"webbot/llm"
	"webbot/trajectory"
)

type LLMActor struct {
	ChatModel            llm.ChatModel
	SystemPrompt         string
	PermissibleFunctions []*llm.FunctionDef
}

//go:embed system_prompt_to_act_on_browser.txt
var systemPromptToActOnBrowser string

func NewLLMActor(chatModel llm.ChatModel) actor.Actor {
	permissibleFunctions := []*llm.FunctionDef{
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
						Description: "The text to send to the user. This function should be called when you want to respond to the user.",
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
	return &LLMActor{
		ChatModel:            chatModel,
		SystemPrompt:         systemPromptToActOnBrowser,
		PermissibleFunctions: permissibleFunctions,
	}
}

func (a *LLMActor) permissibleFunctionMap() map[string]*llm.FunctionDef {
	m := make(map[string]*llm.FunctionDef)
	for _, functionDef := range a.PermissibleFunctions {
		m[functionDef.Name] = functionDef
	}
	return m
}

const maxTokenContextWindowMarginProportion float32 = 0.1

func (a *LLMActor) NextAction(ctx context.Context, traj *trajectory.Trajectory, br *browser.Browser) (nextAction trajectory.TrajectoryItem, render trajectory.TrajectoryItem, err error) {
	_, pageRender, err := br.Render(language.LanguageMD)
	if err != nil {
		return nil, nil, fmt.Errorf("browser failed to render page: %w", err)
	}
	state := fmt.Sprintf(`----- START BROWSER -----
%s
----- END BROWSER -----

----- START TRAJECTORY -----
%s
----- END TRAJECTORY -----
`, pageRender, traj.GetAbbreviatedText())
	messages := []*llm.Message{
		{
			Role:    llm.MessageRoleSystem,
			Content: a.SystemPrompt,
		},
		{
			Role:    llm.MessageRoleUser,
			Content: state,
		},
	}
	var messageDebugDisplayItem trajectory.TrajectoryItem
	if messageDebugDisplay, err := a.renderDebugDisplay(messages, a.PermissibleFunctions); err != nil {
		messageDebugDisplayItem = trajectory.NewDebugRenderedDisplay(trajectory.DebugDisplayTypeLLMMesages, fmt.Sprintf("error rendering message debug display: %s", err))
	} else {
		messageDebugDisplayItem = trajectory.NewDebugRenderedDisplay(trajectory.DebugDisplayTypeLLMMesages, messageDebugDisplay)
	}
	approxNumTokens := llm.ApproxNumTokensInMessages(messages) + llm.ApproxNumTokensInFunctionDefs(a.PermissibleFunctions)
	if approxNumTokens > int(float32(a.ChatModel.ContextLength())*(1-maxTokenContextWindowMarginProportion)) {
		return trajectory.NewErrorMaxContextLengthExceeded(a.ChatModel.ContextLength(), approxNumTokens), messageDebugDisplayItem, nil
	}
	if res, err := a.ChatModel.Message(ctx, messages, &llm.MessageOptions{
		Temperature: 0.0,
		Functions:   a.PermissibleFunctions,
	}); err != nil {
		return nil, nil, fmt.Errorf("failed to generate message: %w", err)
	} else if res.FunctionCall == nil {
		return trajectory.NewAgentMessage(res.Content), messageDebugDisplayItem, nil
	} else if _, ok := a.permissibleFunctionMap()[res.FunctionCall.Name]; !ok {
		return nil, nil, fmt.Errorf("unsupported action was attempted: %s", res.FunctionCall.Name)
	} else if nextAction, err := a.parseNextAction(res.FunctionCall.Name, res.FunctionCall.Arguments); err != nil {
		return nil, nil, fmt.Errorf("failed to parse next action: %w", err)
	} else {
		return nextAction, messageDebugDisplayItem, nil
	}
}

func (a *LLMActor) renderDebugDisplay(messages []*llm.Message, functionDefs []*llm.FunctionDef) (string, error) {
	messageTexts := make([]string, len(messages))
	for i, message := range messages {
		if message.Role == llm.MessageRoleFunction {
			messageTexts[i] = fmt.Sprintf("%s(%s)", message.FunctionCall.Name, message.FunctionCall.Arguments)
		} else {
			messageTexts[i] = fmt.Sprintf("%s: %s", message.Role, message.Content)
		}
	}

	functionDefTexts := make([]string, len(functionDefs))
	for i, functionDef := range functionDefs {
		b, err := json.Marshal(functionDef)
		if err != nil {
			return "", fmt.Errorf("error marshaling function def: %w", err)
		}
		functionDefTexts[i] = string(b)
	}
	return fmt.Sprintf("%s\n\n%s", strings.Join(messageTexts, "\n"), strings.Join(functionDefTexts, "\n")), nil
}

func (a *LLMActor) parseNextAction(name string, arguments string) (trajectory.TrajectoryItem, error) {
	var args map[string]any
	functions := a.permissibleFunctionMap()
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
