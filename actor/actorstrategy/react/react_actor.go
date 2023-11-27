package react

import (
	"context"
	"webbot/actor/actorstrategy"
	"webbot/browser"
	"webbot/llm"
	"webbot/trajectory"
)

type ReactActor struct {
	models *llm.Models
}

func New(models *llm.Models) actorstrategy.ActorStrategy {
	return &ReactActor{
		models: models,
	}
}

func (ra *ReactActor) NextAction(ctx context.Context, traj *trajectory.Trajectory, br *browser.Browser) (trajectory.TrajectoryItem, error) {
	// TODO: implement
	return nil, nil
}
