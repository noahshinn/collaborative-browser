package trajectory

import (
	"fmt"
	"webbot/browser/virtualid"
)

type TrajectoryItem interface {
	GetText() string
}

type BrowserAction struct {
	Type BrowserActionType   `json:"type"`
	ID   virtualid.VirtualID `json:"id"`

	// for send_keys
	Text string `json:"text"`

	// for navigate
	URL string `json:"url"`
}

func NewBrowserClickAction(id virtualid.VirtualID) *BrowserAction {
	return &BrowserAction{
		Type: BrowserActionTypeClick,
		ID:   id,
	}
}

func NewBrowserSendKeysAction(id virtualid.VirtualID, text string) *BrowserAction {
	return &BrowserAction{
		Type: BrowserActionTypeSendKeys,
		ID:   id,
		Text: text,
	}
}

func NewBrowserNavigateAction(url string) *BrowserAction {
	return &BrowserAction{
		Type: BrowserActionTypeNavigate,
		URL:  url,
	}
}

func (ba *BrowserAction) GetText() string {
	switch ba.Type {
	case BrowserActionTypeClick:
		return fmt.Sprintf("%s(id=%s)", ba.Type, ba.ID)
	case BrowserActionTypeSendKeys:
		return fmt.Sprintf("%s(id=%s, text=%s)", ba.Type, ba.ID, ba.Text)
	case BrowserActionTypeNavigate:
		return fmt.Sprintf("%s(url=%s)", ba.Type, ba.URL)
	default:
		panic(fmt.Sprintf("unsupported browser action type: %s", ba.Type))
	}
}

type BrowserActionType string

const (
	BrowserActionTypeClick    BrowserActionType = "click"
	BrowserActionTypeSendKeys BrowserActionType = "send_keys"
	BrowserActionTypeNavigate BrowserActionType = "navigate"
)

type BrowserObservation struct {
	Display string
}

func NewBrowserObservation(display string) *BrowserObservation {
	return &BrowserObservation{
		Display: display,
	}
}

func (bo *BrowserObservation) GetText() string {
	return bo.Display
}
