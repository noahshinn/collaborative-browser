package verificationactor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"webbot/actor"
	"webbot/actor/llmactor"
	"webbot/llm"
	"webbot/utils/jsonx"
)

type VerificationActor struct {
	model             llm.ChatModel
	verificationModel llm.ChatModel
	llmActor          actor.StringActor
}

type Options struct {
	VerificationModel llm.ChatModel
	LLMActor          actor.StringActor
}

func NewVerificationActor(model llm.ChatModel, options *Options) actor.StringActor {
	verificationModel := model
	llmActor := llmactor.NewLLMActor(model)
	if options != nil {
		if options.VerificationModel != nil {
			verificationModel = options.VerificationModel
		}
		if options.LLMActor != nil {
			llmActor = options.LLMActor
		}
	}
	return &VerificationActor{
		model:             model,
		verificationModel: verificationModel,
		llmActor:          llmActor,
	}
}

// Iterative:
//   - 1. Sample an action a_0 from the model
//   - 2. Sample {0, 1} from the verification model
//   - 3. If 0, sample from the action space without a_0
//   - 4. Continue until the verification model returns 1
func (va *VerificationActor) Act(ctx context.Context, messages []*llm.Message, functionDefs []*llm.FunctionDef) (string, error) {
	var actionSpace []*llm.FunctionDef
	copy(actionSpace, functionDefs)
	for len(actionSpace) > 0 {
		nextAction, err := va.llmActor.Act(ctx, messages, actionSpace)
		if err != nil {
			return "", fmt.Errorf("error acting with llm actor: %w", err)
		}
		actionSpaceReprs := map[string]string{}
		for _, actionDef := range actionSpace {
			actionDefRepr, err := jsonx.StructToString(actionDef)
			if err != nil {
				return "", fmt.Errorf("error marshaling action def: %w", err)
			}
			actionSpaceReprs[actionDef.Name] = actionDefRepr
		}
		if verificationDecision, err := va.Verify(ctx, "", nextAction, []string{}); err != nil {
			return "", fmt.Errorf("error verifying action: %w", err)
		} else if verificationDecision {
			return nextAction, nil
		}
		newActionSpace := []*llm.FunctionDef{}
		for _, actionDef := range actionSpace {
			if actionDef.Name != nextAction {
				newActionSpace = append(newActionSpace, actionDef)
			}
		}
		if len(newActionSpace) == len(actionSpace) {
			return "", fmt.Errorf("next action %s was not in action space", nextAction)
		}
		actionSpace = newActionSpace
	}
	// TODO: possibly invoke a fallback action here
	return "", fmt.Errorf("could not find a valid action")
}

// TODO: implement
const verificationSystemPrompt = "Determine if the proposed Action is correct."

func (va *VerificationActor) Verify(ctx context.Context, state string, actionRepr string, actionSpaceReprs []string) (bool, error) {
	messages := []*llm.Message{
		{
			Role:    llm.MessageRoleSystem,
			Content: verificationSystemPrompt,
		},
		{
			Role:    llm.MessageRoleUser,
			Content: fmt.Sprintf("# State:\n%s\n\n# Action Space:\n%s\n\nProposed Action:\n%s", state, strings.Join(actionSpaceReprs, "\n"), actionRepr),
		},
	}
	verificationFunctionDef := &llm.FunctionDef{
		Name: "verification",
		Parameters: llm.Parameters{
			Type: "object",
			Properties: map[string]llm.Property{
				"verification": {
					Type:        "bool",
					Description: "Whether the action is correct or not",
				},
			},
			Required: []string{"verification"},
		},
	}
	if res, err := va.verificationModel.Message(ctx, messages, &llm.MessageOptions{
		Temperature:  0.0,
		Functions:    []*llm.FunctionDef{verificationFunctionDef},
		FunctionCall: "verification",
	}); err != nil {
		return false, err
	} else if res.FunctionCall == nil {
		return false, fmt.Errorf("verification model did not return a function call")
	} else if res.FunctionCall.Name != "verification" {
		return false, fmt.Errorf("verification model returned a function call with name %s instead of verification", res.FunctionCall.Name)
	} else {
		var args map[string]any
		if err := json.Unmarshal([]byte(res.FunctionCall.Arguments), &args); err != nil {
			return false, fmt.Errorf("error unmarshaling verification function call arguments: %w", err)
		} else if _, ok := args["verification"]; !ok {
			return false, fmt.Errorf("verification function call arguments did not contain verification")
		} else if verificationDecision, ok := args["verification"].(bool); !ok {
			return false, fmt.Errorf("verification function call argument verification was not a bool")
		} else {
			return verificationDecision, nil
		}
	}
}
