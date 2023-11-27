package functionafforder

import (
	"collaborativebrowser/afforder/afforderstrategy"
	"collaborativebrowser/browser"
	"collaborativebrowser/browser/language"
	"collaborativebrowser/browser/virtualid"
	"collaborativebrowser/llm"
	"collaborativebrowser/trajectory"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

type FunctionAfforder struct {
	permissibleFunctions   []*llm.FunctionDef
	permissibleFunctionMap map[string]*llm.FunctionDef
}

//go:embed system_prompt_to_act_on_browser.txt
var systemPromptToActOnBrowser string

func New() afforderstrategy.AfforderStrategy {
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
	m := make(map[string]*llm.FunctionDef)
	for _, functionDef := range permissibleFunctions {
		m[functionDef.Name] = functionDef
	}
	return &FunctionAfforder{
		permissibleFunctions:   permissibleFunctions,
		permissibleFunctionMap: m,
	}
}

func (a *FunctionAfforder) GetAffordances(traj *trajectory.Trajectory, br *browser.Browser) ([]*llm.Message, []*llm.FunctionDef, error) {
	pageRender, err := br.Render(language.LanguageMD)
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
			Content: systemPromptToActOnBrowser,
		},
		{
			Role:    llm.MessageRoleUser,
			Content: fmt.Sprintf("%s\n\nLook at the Trajectory to inform your next action.", strings.TrimSpace(state)),
		},
	}
	return messages, a.permissibleFunctions, nil
}

func (a *FunctionAfforder) ParseNextAction(name string, arguments string) (trajectory.TrajectoryItem, error) {
	var args map[string]any
	functions := a.permissibleFunctionMap
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		if name == "message" && len(arguments) > 0 {
			return trajectory.NewAgentMessage(arguments), nil
		}
		return nil, fmt.Errorf("error unmarshaling function call arguments for function %s: %w, \"%s\"", name, err, arguments)
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
	case "task_not_possible":
		return trajectory.NewBrowserTaskNotPossibleAction(args["reason"].(string)), nil
	default:
		return nil, fmt.Errorf("unsupported action was attempted: %s", name)
	}
}

func (a *FunctionAfforder) DoesActionExist(name string) bool {
	_, ok := a.permissibleFunctionMap[name]
	return ok
}
