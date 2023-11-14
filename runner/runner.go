package runner

import (
	"webbot/afforder"
	"webbot/browser"
	"webbot/trajectory"
)

type Runner interface {
	Run() error
	RunAndStream() (<-chan *trajectory.TrajectoryStreamEvent, error)
	Afforder() afforder.Afforder
	Trajectory() *trajectory.Trajectory
	Browser() *browser.Browser
	Log() error
}
