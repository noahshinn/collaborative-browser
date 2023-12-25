package finiterunner

import (
	"collaborativebrowser/trajectory"
	"fmt"
)

func (r *FiniteRunner) sharedTrajectoryAddItem(item trajectory.TrajectoryItem) error {
	return r.sharedTrajectoryAddItems([]trajectory.TrajectoryItem{item})
}

func (r *FiniteRunner) sharedTrajectoryAddItems(items []trajectory.TrajectoryItem) error {
	traj, err := r.sharedTrajectoryRead()
	if err != nil {
		traj = &trajectory.Trajectory{}
	}
	traj.AddItems(items)
	trajJSONString, err := trajectory.MarshalTrajectory(traj)
	if err != nil {
		return fmt.Errorf("failed to marshal shared trajectory: %w", err)
	}
	r.sharedLocalStorage.Set("traj", string(trajJSONString))
	return nil
}

func (r *FiniteRunner) sharedTrajectoryRead() (*trajectory.Trajectory, error) {
	if trajJSONString, err := r.sharedLocalStorage.Get("traj"); err != nil {
		return nil, fmt.Errorf("error reading traj from shared local storage: %w", err)
	} else {
		return trajectory.UnmarshalTrajectory([]byte(trajJSONString))
	}
}
