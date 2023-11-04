package trajectory

import (
	"fmt"
	"strings"
	"webbot/browser/virtualid"
	"webbot/utils/slicesx"
)

type Trajectory struct {
	Items []TrajectoryItem
}

func (t *Trajectory) GetText() string {
	if len(t.Items) == 0 {
		return ""
	}
	itemTexts := slicesx.Map(t.Items[:len(t.Items)-1], func(item TrajectoryItem) string {
		return item.GetText()
	})
	lastItem := t.Items[len(t.Items)-1].GetText()
	return fmt.Sprintf("Trajectory:\n%s", strings.Join(append(itemTexts, lastItem), "\n"))
}

func (t *Trajectory) AddItem(item TrajectoryItem) {
	t.Items = append(t.Items, item)
}

type TrajectoryItem interface {
	GetAbbreviatedText() string
	GetText() string
	ShouldHandoff() bool
}

type TrajectoryStreamEvent struct {
	TrajectoryItem TrajectoryItem
	Error          error
}

type BrowserAction struct {
	Type BrowserActionType   `json:"type"`
	ID   virtualid.VirtualID `json:"id"`

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
		Type: BrowserActionTypeClick,
		ID:   id,
	}
}

func NewBrowserSendKeysAction(id virtualid.VirtualID, text string) TrajectoryItem {
	return &BrowserAction{
		Type: BrowserActionTypeSendKeys,
		ID:   id,
		Text: text,
	}
}

func NewBrowserNavigateAction(url string) TrajectoryItem {
	return &BrowserAction{
		Type: BrowserActionTypeNavigate,
		URL:  url,
	}
}

func NewBrowserTaskCompleteAction(reason string) TrajectoryItem {
	return &BrowserAction{
		Type:   BrowserActionTypeTaskComplete,
		Reason: reason,
	}
}

func NewBrowserTaskNotPossibleAction(reason string) TrajectoryItem {
	return &BrowserAction{
		Type:   BrowserActionTypeTaskNotPossible,
		Reason: reason,
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
	case BrowserActionTypeTaskComplete:
		return fmt.Sprintf("%s(reason=%s)", ba.Type, ba.Reason)
	case BrowserActionTypeTaskNotPossible:
		return fmt.Sprintf("%s(reason=%s)", ba.Type, ba.Reason)
	default:
		panic(fmt.Sprintf("unsupported browser action type: %s", ba.Type))
	}
}

func (ba *BrowserAction) GetAbbreviatedText() string {
	// there may be room to truncate some action types this
	return ba.GetText()
}

func (ba *BrowserAction) ShouldHandoff() bool {
	return ba.Type == BrowserActionTypeTaskComplete || ba.Type == BrowserActionTypeTaskNotPossible
}

type BrowserObservation struct {
	Text            string
	TextAbbreviated string
}

func NewBrowserObservation(text string) TrajectoryItem {
	return &BrowserObservation{
		Text: text,
	}
}

func (bo *BrowserObservation) GetText() string {
	return bo.Text
}

func (bo *BrowserObservation) GetAbbreviatedText() string {
	return bo.TextAbbreviated
}

func (bo *BrowserObservation) ShouldHandoff() bool {
	return false
}

type Message struct {
	Author MessageAuthor
	Text   string
}

type MessageAuthor string

const (
	MessageAuthorUser  MessageAuthor = "user"
	MessageAuthorAgent MessageAuthor = "agent"
)

func NewUserMessage(text string) TrajectoryItem {
	return &Message{
		Author: MessageAuthorUser,
		Text:   text,
	}
}

func NewAgentMessage(text string) TrajectoryItem {
	return &Message{
		Author: MessageAuthorAgent,
		Text:   text,
	}
}

func (m *Message) GetText() string {
	return fmt.Sprintf("%s: %s", m.Author, m.Text)
}

const DefaultAgentMessageAbbreviationLength = 100

func (m *Message) GetAbbreviatedText() string {
	text := m.GetText()
	if m.Author == MessageAuthorAgent && len(m.Text) > DefaultAgentMessageAbbreviationLength {
		return fmt.Sprintf("%s...", text[:DefaultAgentMessageAbbreviationLength])
	}
	return text
}

func (m *Message) ShouldHandoff() bool {
	return true
}

type ErrorMaxNumStepsReached struct {
	MaxNumSteps int
}

func NewErrorMaxNumStepsReached(maxNumSteps int) TrajectoryItem {
	return &ErrorMaxNumStepsReached{
		MaxNumSteps: maxNumSteps,
	}
}

func (m *ErrorMaxNumStepsReached) GetText() string {
	return fmt.Sprintf("max num steps reached: %d", m.MaxNumSteps)
}

func (m *ErrorMaxNumStepsReached) GetAbbreviatedText() string {
	return m.GetText()
}

func (m *ErrorMaxNumStepsReached) ShouldHandoff() bool {
	return true
}
