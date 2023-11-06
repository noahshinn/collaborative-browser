package afforder

import (
	"context"
	"webbot/browser"
	"webbot/trajectory"
)

type Afforder interface {
	NextAction(ctx context.Context, traj *trajectory.Trajectory, br *browser.Browser) (nextAction trajectory.TrajectoryItem, render trajectory.TrajectoryItem, err error)
}
