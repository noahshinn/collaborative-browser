package trajectory

import (
	"strings"
)

type Trajectory struct {
	Items []*TrajectoryItem
}

func (t *Trajectory) GetText() string {
	if len(t.Items) == 0 {
		return ""
	}
	itemTexts := []string{}
	for _, item := range t.Items {
		if item.ShouldRender {
			itemTexts = append(itemTexts, item.GetText())
		}
	}
	return strings.Join(itemTexts, "\n")
}

func (t *Trajectory) GetAbbreviatedText() string {
	if len(t.Items) == 0 {
		return ""
	}
	itemTexts := []string{}
	for _, item := range t.Items {
		if item.ShouldRender {
			if item.Type == TrajectoryItemMessage {
				itemTexts = append(itemTexts, item.GetText())
			} else {
				itemTexts = append(itemTexts, item.GetAbbreviatedText())
			}
		}
	}
	return strings.Join(itemTexts, "\n")
}

type TrajectoryItem struct {
	Type          TrajectoryItemType `json:"type"`
	ShouldHandoff bool               `json:"should_handoff"`
	ShouldRender  bool               `json:"should_render"`

	// for `click`, `send_keys`
	ID string `json:"id"`

	// for `navigate`
	URL string `json:"url"`

	// for `send_keys` and `message`
	Text string `json:"text"`

	// for `task_complete`, `task_not_possible`
	Reason string `json:"reason"`

	// for `max_num_steps_reached`
	MaxNumSteps int `json:"max_num_steps"`

	// for `max_context_length_exceeded`
	ContextLengthAllowed  int `json:"context_length_allowed"`
	ContextLengthReceived int `json:"context_length_received"`

	// for `message`
	Author MessageAuthor `json:"author"`
}

type TrajectoryItemType string

const (
	TrajectoryItemMessage                  TrajectoryItemType = "message"
	TrajectoryItemObservation              TrajectoryItemType = "browser_observation"
	TrajectoryItemMaxNumStepsReached       TrajectoryItemType = "max_num_steps_reached"
	TrajectoryItemMaxContextLengthExceeded TrajectoryItemType = "max_context_length_exceeded"
	TrajectoryItemClick                    TrajectoryItemType = "click"
	TrajectoryItemSendKeys                 TrajectoryItemType = "send_keys"
	TrajectoryItemNavigate                 TrajectoryItemType = "navigate"
	TrajectoryItemTaskComplete             TrajectoryItemType = "task_complete"
	TrajectoryItemTaskNotPossible          TrajectoryItemType = "task_not_possible"
)

type TrajectoryStreamEvent struct {
	TrajectoryItem *TrajectoryItem
	Error          error
}
