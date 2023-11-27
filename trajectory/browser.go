package trajectory

import (
	"collaborativebrowser/browser/virtualid"
	"fmt"
)

type BrowserAction struct {
	Type   BrowserActionType   `json:"type"`
	ID     virtualid.VirtualID `json:"id"`
	Render bool                `json:"render"`

	// for send_keys
	Text string `json:"text"`

	// for navigate
	URL string `json:"url"`

	// for task_complete or task_not_possible
	Reason string `json:"reason"`
}

type BrowserActionType string

const (
	BrowserActionTypeClick           BrowserActionType = "click"
	BrowserActionTypeSendKeys        BrowserActionType = "send_keys"
	BrowserActionTypeNavigate        BrowserActionType = "navigate"
	BrowserActionTypeTaskComplete    BrowserActionType = "task_complete"
	BrowserActionTypeTaskNotPossible BrowserActionType = "task_not_possible"
)

func NewBrowserClickAction(id virtualid.VirtualID) TrajectoryItem {
	return &BrowserAction{
		Type:   BrowserActionTypeClick,
		ID:     id,
		Render: true,
	}
}

func NewBrowserSendKeysAction(id virtualid.VirtualID, text string) TrajectoryItem {
	return &BrowserAction{
		Type:   BrowserActionTypeSendKeys,
		ID:     id,
		Text:   text,
		Render: true,
	}
}

func NewBrowserNavigateAction(url string) TrajectoryItem {
	return &BrowserAction{
		Type:   BrowserActionTypeNavigate,
		URL:    url,
		Render: true,
	}
}

func NewBrowserTaskCompleteAction(reason string) TrajectoryItem {
	return &BrowserAction{
		Type:   BrowserActionTypeTaskComplete,
		Reason: reason,
		Render: true,
	}
}

func NewBrowserTaskNotPossibleAction(reason string) TrajectoryItem {
	return &BrowserAction{
		Type:   BrowserActionTypeTaskNotPossible,
		Reason: reason,
		Render: true,
	}
}

func (ba *BrowserAction) GetText() string {
	var text string
	switch ba.Type {
	case BrowserActionTypeClick:
		text = fmt.Sprintf("%s(id=%s)", ba.Type, ba.ID)
	case BrowserActionTypeSendKeys:
		text = fmt.Sprintf("%s(id=%s, text=\"%s\")", ba.Type, ba.ID, ba.Text)
	case BrowserActionTypeNavigate:
		text = fmt.Sprintf("%s(url=\"%s\")", ba.Type, ba.URL)
	case BrowserActionTypeTaskComplete:
		text = fmt.Sprintf("%s(reason=\"%s\")", ba.Type, ba.Reason)
	case BrowserActionTypeTaskNotPossible:
		text = fmt.Sprintf("%s(reason=\"%s\")", ba.Type, ba.Reason)
	default:
		panic(fmt.Sprintf("unsupported browser action type: %s", ba.Type))
	}
	return fmt.Sprintf("action: %s", text)
}

func (ba *BrowserAction) GetAbbreviatedText() string {
	// there may be room to truncate some action types
	return ba.GetText()
}

func (ba *BrowserAction) ShouldHandoff() bool {
	return ba.Type == BrowserActionTypeTaskComplete || ba.Type == BrowserActionTypeTaskNotPossible
}

func (ba *BrowserAction) ShouldRender() bool {
	return ba.Render
}

type BrowserObservation struct {
	DontHandoff
	Render
	Text            string
	TextAbbreviated string
}

func NewBrowserObservation(text string) TrajectoryItem {
	return &BrowserObservation{
		Text: text,
	}
}

func (bo *BrowserObservation) GetText() string {
	return fmt.Sprintf("observation: %s", bo.Text)
}

func (bo *BrowserObservation) GetAbbreviatedText() string {
	text := bo.Text
	if len(text) > 100 {
		text = text[:100] + "..."
	}
	return fmt.Sprintf("observation: %s", text)
}
