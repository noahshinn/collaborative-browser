package basellm

import (
	"collaborativebrowser/actor/actorstrategy"
	"collaborativebrowser/afforder"
	"collaborativebrowser/afforder/afforderstrategy"
	"collaborativebrowser/browser"
	"collaborativebrowser/llm"
	"collaborativebrowser/trajectory"
	"context"
	"fmt"
)

type BaseLLMActor struct {
	models   *llm.Models
	afforder afforderstrategy.AfforderStrategy
}

func New(models *llm.Models, options *actorstrategy.Options) actorstrategy.ActorStrategy {
	afforderStrategyID := afforder.DefaultAfforderStrategyID
	if options != nil {
		if options.AfforderStrategyID != "" {
			afforderStrategyID = options.AfforderStrategyID
		}
	}
	a := afforder.AfforderStrategyByID(afforderStrategyID, models)
	return &BaseLLMActor{
		models:   models,
		afforder: a,
	}
}

func (a *BaseLLMActor) NextAction(ctx context.Context, traj *trajectory.Trajectory, br *browser.Browser) (trajectory.TrajectoryItem, error) {
	messages, functionDefs, err := a.afforder.GetAffordances(ctx, traj, br)
	if err != nil {
		return nil, fmt.Errorf("failed to get affordances: %w", err)
	}
	if res, err := a.models.DefaultChatModel.Message(ctx, messages, &llm.MessageOptions{
		Temperature: 0.0,
		Functions:   functionDefs,
	}); err != nil {
		return nil, fmt.Errorf("failed to generate message: %w", err)
	} else if res.FunctionCall == nil {
		return trajectory.NewMessage(trajectory.MessageAuthorAgent, res.Content), nil
	} else if exists := a.afforder.DoesActionExist(res.FunctionCall.Name); !exists {
		return nil, fmt.Errorf("unsupported action was attempted: %s", res.FunctionCall.Name)
	} else if nextAction, err := a.afforder.ParseNextAction(res.FunctionCall.Name, res.FunctionCall.Arguments); err != nil {
		return nil, fmt.Errorf("failed to parse next action: %w", err)
	} else {
		return nextAction, nil
	}
}
