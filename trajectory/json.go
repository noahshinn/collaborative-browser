package trajectory

import (
	"encoding/json"
	"fmt"
)

type TrajectoryItemJSON struct {
	Type TrajectoryItemType `json:"type"`
	Data json.RawMessage    `json:"data"`
}

type TrajectoryItemType string

const (
	TrajectoryItemTypeMessage                  TrajectoryItemType = "message"
	TrajectoryItemTypeBrowserObservation       TrajectoryItemType = "browser_observation"
	TrajectoryItemTypeBrowserAction            TrajectoryItemType = "browser_action"
	TrajectoryItemTypeMaxNumStepsReached       TrajectoryItemType = "max_num_steps_reached"
	TrajectoryItemTypeMaxContextLengthExceeded TrajectoryItemType = "max_context_length_exceeded"
)

func MarshalTrajectory(traj *Trajectory) ([]byte, error) {
	var trajJSON []*TrajectoryItemJSON
	for _, item := range traj.Items {
		itemJSON, err := TrajectoryItemToJSON(item)
		if err != nil {
			return nil, err
		}
		trajJSON = append(trajJSON, itemJSON)
	}
	return json.Marshal(trajJSON)
}

func UnmarshalTrajectory(data []byte) (*Trajectory, error) {
	var trajJSON []*TrajectoryItemJSON
	err := json.Unmarshal(data, &trajJSON)
	if err != nil {
		return nil, err
	}
	traj := &Trajectory{Items: []TrajectoryItem{}}
	for _, itemJSON := range trajJSON {
		item, err := JSONToTrajectoryItem(itemJSON)
		if err != nil {
			return nil, err
		}
		traj.Items = append(traj.Items, item)
	}
	return traj, nil
}

func TrajectoryItemToJSON(item TrajectoryItem) (*TrajectoryItemJSON, error) {
	var typ TrajectoryItemType
	switch item.(type) {
	case *Message:
		typ = TrajectoryItemTypeMessage
	case *BrowserObservation:
		typ = TrajectoryItemTypeBrowserObservation
	case *BrowserAction:
		typ = TrajectoryItemTypeBrowserAction
	case *ErrorMaxNumStepsReached:
		typ = TrajectoryItemTypeMaxNumStepsReached
	case *ErrorMaxContextLengthExceeded:
		typ = TrajectoryItemTypeMaxContextLengthExceeded
	default:
		return nil, fmt.Errorf("unknown trajectory item type: %T", item)
	}
	data, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	return &TrajectoryItemJSON{
		Type: typ,
		Data: data,
	}, nil
}

func MarshalTrajectoryItem(item TrajectoryItem) ([]byte, error) {
	itemJSON, err := TrajectoryItemToJSON(item)
	if err != nil {
		return nil, err
	}
	return json.Marshal(itemJSON)
}

func UnmarshalTrajectoryItem(data []byte) (TrajectoryItem, error) {
	var trajItemJSON *TrajectoryItemJSON
	err := json.Unmarshal(data, &trajItemJSON)
	if err != nil {
		return nil, err
	}
	return JSONToTrajectoryItem(trajItemJSON)
}

func JSONToTrajectoryItem(item *TrajectoryItemJSON) (TrajectoryItem, error) {
	var trajItem TrajectoryItem
	switch TrajectoryItemType(item.Type) {
	case TrajectoryItemTypeMessage:
		trajItem = &Message{}
	case TrajectoryItemTypeBrowserObservation:
		trajItem = &BrowserObservation{}
	case TrajectoryItemTypeBrowserAction:
		trajItem = &BrowserAction{}
	case TrajectoryItemTypeMaxNumStepsReached:
		trajItem = &ErrorMaxNumStepsReached{}
	case TrajectoryItemTypeMaxContextLengthExceeded:
		trajItem = &ErrorMaxContextLengthExceeded{}
	default:
		return nil, fmt.Errorf("unknown trajectory item type: %s", item.Type)
	}
	err := json.Unmarshal(item.Data, trajItem)
	if err != nil {
		return nil, err
	}
	return trajItem, nil
}
