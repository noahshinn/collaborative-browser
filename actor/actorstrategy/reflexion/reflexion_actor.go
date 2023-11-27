package reflexion

import (
	"context"
	"fmt"
	"webbot/actor/actorstrategy"
	"webbot/browser"
	"webbot/llm"
	"webbot/trajectory"
)

type ReflexionActor struct {
	models            *llm.Models
	maxNumIterations  int
	baseActorStrategy actorstrategy.ActorStrategy
}

const DefaultMaxNumIterations = 3

func New(models *llm.Models, baseActorStrategy actorstrategy.ActorStrategy, maxNumIterations int) actorstrategy.ActorStrategy {
	return &ReflexionActor{
		models:            models,
		maxNumIterations:  maxNumIterations,
		baseActorStrategy: baseActorStrategy,
	}
}

func (a *ReflexionActor) NextAction(ctx context.Context, traj *trajectory.Trajectory, br *browser.Browser) (trajectory.TrajectoryItem, error) {
	nextAction, err := a.baseActorStrategy.NextAction(ctx, traj, br)
	if err != nil {
		return nil, fmt.Errorf("error acting with llm actor: %w", err)
	}
	// TODO: do reflection in a loop or once
	return nextAction, nil
}
