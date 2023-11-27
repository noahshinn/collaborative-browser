package afforderstrategy

import (
	"collaborativebrowser/browser"
	"collaborativebrowser/llm"
	"collaborativebrowser/trajectory"
)

type AfforderStrategy interface {
	GetAffordances(traj *trajectory.Trajectory, br *browser.Browser) (messages []*llm.Message, functionDefs []*llm.FunctionDef, err error)
	ParseNextAction(name string, arguments string) (trajectory.TrajectoryItem, error)
	DoesActionExist(name string) bool
}