package runner

import (
	"webbot/actor/actorstrategy"
	"webbot/browser"
	"webbot/trajectory"
)

type Runner interface {
	Run() error
	RunAndStream() (<-chan *trajectory.TrajectoryStreamEvent, error)
	Actor() actorstrategy.ActorStrategy
	Trajectory() *trajectory.Trajectory
	Browser() *browser.Browser
	Log() error
}
