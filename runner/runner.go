package runner

import (
	"collaborativebrowser/actor/actorstrategy"
	"collaborativebrowser/browser"
	"collaborativebrowser/trajectory"
)

type Runner interface {
	Run() error
	RunAndStream() (<-chan *trajectory.TrajectoryStreamEvent, error)
	Actor() actorstrategy.ActorStrategy
	Trajectory() *trajectory.Trajectory
	Browser() *browser.Browser
	Log() error
}
