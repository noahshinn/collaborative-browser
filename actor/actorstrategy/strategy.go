package actorstrategy

import (
	"collaborativebrowser/afforder"
	"collaborativebrowser/browser"
	"collaborativebrowser/trajectory"
	"context"
)

type ActorStrategy interface {
	NextAction(ctx context.Context, traj *trajectory.Trajectory, br *browser.Browser) (*trajectory.TrajectoryItem, error)
}

type Options struct {
	// for basellm
	AfforderStrategyID afforder.AfforderStrategyID

	// for reflexion
	MaxNumIterations int
}
