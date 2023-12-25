package runner

import (
	"collaborativebrowser/trajectory"
)

type Runner interface {
	Run() error
	RunAndStream() (<-chan *trajectory.TrajectoryStreamEvent, error)
	DisplayTrajectory() error
	AddItemToTrajectory(item *trajectory.TrajectoryItem) error
	RunHeadful() error
	RunHeadless() error
	Log() error
	Terminate()
}
