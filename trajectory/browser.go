package trajectory

import (
	"fmt"
)

func NewBrowserClickAction(id string) *TrajectoryItem {
	return &TrajectoryItem{
		Type:          TrajectoryItemClick,
		ID:            id,
		ShouldRender:  true,
		ShouldHandoff: false,
	}
}

func NewBrowserSendKeysAction(id string, text string) *TrajectoryItem {
	return &TrajectoryItem{
		Type:          TrajectoryItemSendKeys,
		ID:            id,
		Text:          text,
		ShouldRender:  true,
		ShouldHandoff: false,
	}
}

func NewBrowserNavigateAction(url string) *TrajectoryItem {
	return &TrajectoryItem{
		Type:          TrajectoryItemNavigate,
		URL:           url,
		ShouldRender:  true,
		ShouldHandoff: false,
	}
}

func NewBrowserTaskCompleteAction(reason string) *TrajectoryItem {
	return &TrajectoryItem{
		Type:          TrajectoryItemTaskComplete,
		Reason:        reason,
		ShouldRender:  true,
		ShouldHandoff: false,
	}
}

func NewBrowserTaskNotPossibleAction(reason string) *TrajectoryItem {
	return &TrajectoryItem{
		Type:          TrajectoryItemTaskNotPossible,
		Reason:        reason,
		ShouldRender:  true,
		ShouldHandoff: false,
	}
}

func (ti *TrajectoryItem) GetText() string {
	var text string
	switch ti.Type {
	case TrajectoryItemClick:
		text = fmt.Sprintf("%s(id=%s)", ti.Type, ti.ID)
	case TrajectoryItemSendKeys:
		text = fmt.Sprintf("%s(id=%s, text=\"%s\")", ti.Type, ti.ID, ti.Text)
	case TrajectoryItemNavigate:
		text = fmt.Sprintf("%s(url=\"%s\")", ti.Type, ti.URL)
	case TrajectoryItemTaskComplete:
		text = fmt.Sprintf("%s(reason=\"%s\")", ti.Type, ti.Reason)
	case TrajectoryItemTaskNotPossible:
		text = fmt.Sprintf("%s(reason=\"%s\")", ti.Type, ti.Reason)
	case TrajectoryItemObservation:
		text = fmt.Sprintf("observation: %s", ti.Text)
	case TrajectoryItemMaxNumStepsReached:
		text = fmt.Sprintf("error: max num steps reached: %d", ti.MaxNumSteps)
	case TrajectoryItemMaxContextLengthExceeded:
		text = fmt.Sprintf("error: max context length exceeded: %d > %d tokens", ti.ContextLengthReceived, ti.ContextLengthAllowed)
	case TrajectoryItemMessage:
		text = fmt.Sprintf("%s: %s", ti.Author, ti.Text)
	default:
		panic(fmt.Sprintf("unsupported browser action type: %s", ti.Type))
	}
	return fmt.Sprintf("action: %s", text)
}

const DefaultAgentMessageAbbreviationLength = 100

func (ti *TrajectoryItem) GetAbbreviatedText() string {
	var text string
	switch ti.Type {
	case TrajectoryItemClick, TrajectoryItemSendKeys, TrajectoryItemNavigate, TrajectoryItemTaskComplete, TrajectoryItemTaskNotPossible, TrajectoryItemMaxNumStepsReached, TrajectoryItemMaxContextLengthExceeded:
		text = ti.GetText()
	case TrajectoryItemObservation:
		text = ti.GetText()
		if len(text) > 100 {
			text = text[:100] + "..."
		}
		text = fmt.Sprintf("observation: %s", text)
	case TrajectoryItemMessage:
		text = ti.GetText()
		if ti.Author == MessageAuthorAgent && len(ti.Text) > DefaultAgentMessageAbbreviationLength {
			text = fmt.Sprintf("%s...", text[:DefaultAgentMessageAbbreviationLength])
		}
	}
	return text
}

func NewBrowserObservation(text string) *TrajectoryItem {
	return &TrajectoryItem{
		Type:          TrajectoryItemObservation,
		Text:          text,
		ShouldHandoff: false,
		ShouldRender:  true,
	}
}
