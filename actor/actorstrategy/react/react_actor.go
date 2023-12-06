package react

import (
	"collaborativebrowser/actor/actorstrategy"
	"collaborativebrowser/afforder"
	"collaborativebrowser/afforder/afforderstrategy"
	"collaborativebrowser/browser"
	"collaborativebrowser/llm"
	"collaborativebrowser/trajectory"
	"context"
)

type ReactActor struct {
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
	return &ReactActor{
		models:   models,
		afforder: a,
	}
}

func (ra *ReactActor) NextAction(ctx context.Context, traj *trajectory.Trajectory, br *browser.Browser) (trajectory.TrajectoryItem, error) {
	// TODO: implement
	return nil, nil
}
