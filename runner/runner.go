package runner

import (
	"collaborativebrowser/trajectory"
)

type Runner interface {
	Run() error
	RunAndStream() (<-chan *trajectory.TrajectoryStreamEvent, error)
	AddItemToTrajectory(item trajectory.TrajectoryItem)
	DisplayTrajectory()
	RunHeadful() error
	RunHeadless() error
	Log() error
	Terminate()
}
