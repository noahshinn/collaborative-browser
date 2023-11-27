package actorstrategy

import (
	"collaborativebrowser/browser"
	"collaborativebrowser/trajectory"
	"context"
)

type ActorStrategy interface {
	NextAction(ctx context.Context, traj *trajectory.Trajectory, br *browser.Browser) (trajectory.TrajectoryItem, error)
}
