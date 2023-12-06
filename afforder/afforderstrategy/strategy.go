package afforderstrategy

import (
	"collaborativebrowser/browser"
	"collaborativebrowser/llm"
	"collaborativebrowser/trajectory"
	"context"
)

type AfforderStrategy interface {
	GetAffordances(ctx context.Context, traj *trajectory.Trajectory, br *browser.Browser) (messages []*llm.Message, functionDefs []*llm.FunctionDef, err error)
	ParseNextAction(name string, arguments string) (trajectory.TrajectoryItem, error)
	DoesActionExist(name string) bool
}
