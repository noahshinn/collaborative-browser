package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"webbot/actor/actorstrategy"
	"webbot/browser"
	"webbot/llm"
	"webbot/trajectory"
)

type VerificationActor struct {
	models            *llm.Models
	baseActorStrategy actorstrategy.ActorStrategy
}

type Options struct {
	BaseActorStrategy actorstrategy.ActorStrategy
}

func New(models *llm.Models, baseActorStrategy actorstrategy.ActorStrategy) actorstrategy.ActorStrategy {
	return &VerificationActor{
		models:            models,
		baseActorStrategy: baseActorStrategy,
	}
}

// Iterative:
//   - 1. Sample an action a_0 from the model
//   - 2. Sample {0, 1} from the verification model
//   - 3. If 0, sample from the action space without a_0
//   - 4. Continue until the verification model returns 1
func (va *VerificationActor) NextAction(ctx context.Context, traj *trajectory.Trajectory, br *browser.Browser) (trajectory.TrajectoryItem, error) {
	// TODO: implement
	return nil, nil
}

// TODO: implement
const verificationSystemPrompt = "Determine if the proposed Action is correct."

func (va *VerificationActor) Verify(ctx context.Context, models *llm.Models, state string, actionRepr string, actionSpaceReprs []string) (bool, error) {
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
	if res, err := va.models.DefaultChatModel.Message(ctx, messages, &llm.MessageOptions{
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
