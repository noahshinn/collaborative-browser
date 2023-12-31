package verification

import (
	"collaborativebrowser/actor/actorstrategy"
	"collaborativebrowser/actor/actorstrategy/basellm"
	"collaborativebrowser/afforder"
	"collaborativebrowser/afforder/afforderstrategy"
	"collaborativebrowser/browser"
	"collaborativebrowser/llm"
	"collaborativebrowser/trajectory"
	"context"
)

type VerificationActor struct {
	models            *llm.Models
	baseActorStrategy actorstrategy.ActorStrategy
	afforder          afforderstrategy.AfforderStrategy
}

func New(models *llm.Models, options *actorstrategy.Options) actorstrategy.ActorStrategy {
	afforderStrategyID := afforder.DefaultAfforderStrategyID
	if options != nil {
		if options.AfforderStrategyID != "" {
			afforderStrategyID = options.AfforderStrategyID
		}
	}
	baseActorStrategy := basellm.New(models, &actorstrategy.Options{
		AfforderStrategyID: afforderStrategyID,
	})
	a := afforder.AfforderStrategyByID(afforderStrategyID, models)
	return &VerificationActor{
		models:            models,
		baseActorStrategy: baseActorStrategy,
		afforder:          a,
	}
}

// Iterative:
//   - 1. Sample an action a_0 from the model
//   - 2. Sample a reward (s, a) -> r \in {0, 1} from the verification model
//   - 3. If r is 0, sample from the action space without a_0
//   - 4. Continue until the verification model returns 1
func (va *VerificationActor) NextAction(ctx context.Context, traj *trajectory.Trajectory, br *browser.Browser) (trajectory.TrajectoryItem, error) {
	// TODO: implement
	return va.baseActorStrategy.NextAction(ctx, traj, br)
}

func (va *VerificationActor) Verify(ctx context.Context, messages []*llm.Message, nextAction trajectory.TrajectoryItem, actionSpace []*llm.FunctionDef) (bool, error) {
	// TODO: implement
	return true, nil
}
