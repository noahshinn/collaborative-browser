package reflexion

import (
	"collaborativebrowser/actor/actorstrategy"
	"collaborativebrowser/actor/actorstrategy/basellm"
	"collaborativebrowser/afforder"
	"collaborativebrowser/afforder/afforderstrategy"
	"collaborativebrowser/browser"
	"collaborativebrowser/llm"
	"collaborativebrowser/trajectory"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
)

type ReflexionActor struct {
	models            *llm.Models
	maxNumIterations  int
	baseActorStrategy actorstrategy.ActorStrategy
	afforder          afforderstrategy.AfforderStrategy
}

const DefaultMaxNumIterations = 3

func New(models *llm.Models, options *actorstrategy.Options) actorstrategy.ActorStrategy {
	afforderStrategyID := afforder.DefaultAfforderStrategyID
	maxNumIterations := DefaultMaxNumIterations
	if options != nil {
		if options.AfforderStrategyID != "" {
			afforderStrategyID = options.AfforderStrategyID
		}
		if options.MaxNumIterations > 0 {
			maxNumIterations = options.MaxNumIterations
		}
	}
	a := afforder.AfforderStrategyByID(afforderStrategyID)
	baseActorStrategy := basellm.New(models, &actorstrategy.Options{
		AfforderStrategyID: afforderStrategyID,
	})
	return &ReflexionActor{
		models:            models,
		maxNumIterations:  maxNumIterations,
		baseActorStrategy: baseActorStrategy,
		afforder:          a,
	}
}

func (a *ReflexionActor) NextAction(ctx context.Context, traj *trajectory.Trajectory, br *browser.Browser) (trajectory.TrajectoryItem, error) {
	_, actionSpace, err := a.afforder.GetAffordances(traj, br)
	if err != nil {
		return nil, fmt.Errorf("failed to get affordances: %w", err)
	}
	var prevInferredAction trajectory.TrajectoryItem
	localTraj := trajectory.Trajectory{
		Items: make([]trajectory.TrajectoryItem, len(traj.Items)),
	}
	copy(localTraj.Items, traj.Items)
	for i := 0; i < a.maxNumIterations; i++ {
		nextAction, err := a.baseActorStrategy.NextAction(ctx, &localTraj, br)
		if prevInferredAction != nil {
			// If there are two consecutive messages, then we assume that the second message is the correct action to take.
			if prevInferredAction.IsMessage() && nextAction.IsMessage() {
				return nextAction, nil
			}
			// If there are two consecutive actions with the same name and arguments, then we assume that the second action is the correct action to take.
			if !prevInferredAction.IsMessage() && !nextAction.IsMessage() && prevInferredAction.GetText() == nextAction.GetText() {
				return nextAction, nil
			}
		}

		if err != nil {
			return nil, fmt.Errorf("failed to get next action from base actor strategy: %w", err)
		}
		if validAction, reflection, err := a.reflect(ctx, &localTraj, nextAction, actionSpace); err != nil {
			return nil, fmt.Errorf("failed to reflection on action \"%s\": %w", nextAction.GetText(), err)
		} else if validAction {
			return nextAction, nil
		} else {
			prevInferredAction = nextAction
			localTraj.Items = append(localTraj.Items, trajectory.NewMessage(
				trajectory.MessageAuthorInternalFeedback,
				fmt.Sprintf("wrong action: %s, feedback: %s", nextAction.GetText(), reflection),
			))
		}
	}
	if prevInferredAction == nil {
		return nil, fmt.Errorf("failed to find a valid action")
	}
	return prevInferredAction, nil
}

//go:embed reflect_system_prompt.txt
var reflectSystemPrompt string

//go:embed truncated_browser_system_prompt.txt
var truncatedBrowserSystemPrompt string

func (a *ReflexionActor) reflect(ctx context.Context, traj *trajectory.Trajectory, nextAction trajectory.TrajectoryItem, actionSpace []*llm.FunctionDef) (validAction bool, reflection string, err error) {
	trajectoryRender := traj.GetText()
	b, err := json.MarshalIndent(actionSpace, "", "  ")
	if err != nil {
		return false, "", fmt.Errorf("failed to marshal action space: %w", err)
	}
	actionSpaceRender := string(b)
	userMessage := fmt.Sprintf(`----- START CONTEXT -----
%s

%s

action space: %s

action choice: %s
----- END CONTEXT -----`, truncatedBrowserSystemPrompt, trajectoryRender, actionSpaceRender, nextAction.GetText())
	reflectMessages := []*llm.Message{
		{
			Role:    llm.MessageRoleSystem,
			Content: reflectSystemPrompt,
		},
		{
			Role:    llm.MessageRoleUser,
			Content: userMessage,
		},
	}
	reflectFunctionDef := &llm.FunctionDef{
		Name:        "action_reward",
		Description: "Determines whether the proposed action is correct or not.",
		Parameters: llm.Parameters{
			Type: "object",
			Properties: map[string]llm.Property{
				"reason": {
					Type:        "string",
					Description: "The reason for the binary classification that will follow",
				},
				"classification": {
					Type:        "boolean",
					Description: "Whether the proposed action is correct or not",
				},
				"correct_action": {
					Type:        "string",
					Description: "[Optional] If the proposed action is incorrect, the correct action to take",
				},
			},
			Required: []string{
				"reason",
				"classification",
			},
		},
	}
	var args map[string]any
	if res, err := a.models.DefaultChatModel.Message(ctx, reflectMessages, &llm.MessageOptions{
		Temperature:  0.0,
		Functions:    []*llm.FunctionDef{reflectFunctionDef},
		FunctionCall: "action_reward",
	}); err != nil {
		return false, "", fmt.Errorf("failed to get reflection from chat model: %w", err)
	} else if res.FunctionCall == nil {
		return false, "", fmt.Errorf("reflection model did not return a function call")
	} else if res.FunctionCall.Name != "action_reward" {
		return false, "", fmt.Errorf("reflection model returned a function call with name %s instead of \"action_reward\"", res.FunctionCall.Name)
	} else if err := json.Unmarshal([]byte(res.FunctionCall.Arguments), &args); err != nil {
		return false, "", fmt.Errorf("error unmarshaling verification function call arguments: %w", err)
	} else if reasonRaw, ok := args["reason"]; !ok {
		return false, "", fmt.Errorf("action_reward function call arguments did not contain \"reason\"")
	} else if reason, ok := reasonRaw.(string); !ok {
		return false, "", fmt.Errorf("action_reward function call argument \"reason\" was not a string")
	} else if classificationRaw, ok := args["classification"]; !ok {
		return false, "", fmt.Errorf("action_reward function call arguments did not contain \"classification\"")
	} else if classification, ok := classificationRaw.(bool); !ok {
		return false, "", fmt.Errorf("action_reward function call argument \"classification\" was not a bool")
	} else {
		return classification, reason, nil
	}
}
