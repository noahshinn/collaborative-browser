package basellm

import (
	"context"
	"fmt"
	"webbot/actor/actorstrategy"
	"webbot/afforder/afforderstrategy"
	"webbot/afforder/afforderstrategy/functionafforder"
	"webbot/browser"
	"webbot/llm"
	"webbot/trajectory"
)

type BaseLLMActor struct {
	models   *llm.Models
	afforder afforderstrategy.AfforderStrategy
}

func New(models *llm.Models) actorstrategy.ActorStrategy {
	a := functionafforder.New()
	return &BaseLLMActor{
		models:   models,
		afforder: a,
	}
}

func (a *BaseLLMActor) NextAction(ctx context.Context, traj *trajectory.Trajectory, br *browser.Browser) (trajectory.TrajectoryItem, error) {
	messages, functionDefs, err := a.afforder.GetAffordances(traj, br)
	if err != nil {
		return nil, fmt.Errorf("failed to get affordances: %w", err)
	}
	if res, err := a.models.DefaultChatModel.Message(ctx, messages, &llm.MessageOptions{
		Temperature: 0.0,
		Functions:   functionDefs,
	}); err != nil {
		return nil, fmt.Errorf("failed to generate message: %w", err)
	} else if res.FunctionCall == nil {
		return trajectory.NewAgentMessage(res.Content), nil
	} else if exists := a.afforder.DoesActionExist(res.FunctionCall.Name); !exists {
		return nil, fmt.Errorf("unsupported action was attempted: %s", res.FunctionCall.Name)
	} else if nextAction, err := a.afforder.ParseNextAction(res.FunctionCall.Name, res.FunctionCall.Arguments); err != nil {
		return nil, fmt.Errorf("failed to parse next action: %w", err)
	} else {
		return nextAction, nil
	}
}
