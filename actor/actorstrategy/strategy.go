package actorstrategy

import (
	"context"
	"webbot/browser"
	"webbot/trajectory"
)

type ActorStrategy interface {
	NextAction(ctx context.Context, traj *trajectory.Trajectory, br *browser.Browser) (trajectory.TrajectoryItem, error)
}
