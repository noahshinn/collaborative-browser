package runner

import (
	act "webbot/actor"
	"webbot/browser"
	"webbot/trajectory"
)

type Runner interface {
	Run() error
	RunAndStream() (<-chan *trajectory.TrajectoryStreamEvent, error)
	Actor() act.Actor
	Trajectory() *trajectory.Trajectory
	Browser() *browser.Browser
	Log(filepath string) error
}
