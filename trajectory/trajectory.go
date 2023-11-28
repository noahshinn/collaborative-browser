package trajectory

import (
	"strings"
)

type Trajectory struct {
	Items []TrajectoryItem
}

func (t *Trajectory) GetText() string {
	if len(t.Items) == 0 {
		return ""
	}
	itemTexts := []string{}
	for _, item := range t.Items {
		if item.ShouldRender() {
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
		if item.ShouldRender() {
			if _, ok := item.(*Message); ok {
				itemTexts = append(itemTexts, item.GetText())
			} else {
				itemTexts = append(itemTexts, item.GetAbbreviatedText())
			}
		}
	}
	return strings.Join(itemTexts, "\n")
}

func (t *Trajectory) AddItem(item TrajectoryItem) {
	t.Items = append(t.Items, item)
}

func (t *Trajectory) AddItems(items []TrajectoryItem) {
	t.Items = append(t.Items, items...)
}

type TrajectoryItem interface {
	GetAbbreviatedText() string
	GetText() string
	ShouldHandoff() bool
	ShouldRender() bool
	IsMessage() bool
}

type TrajectoryStreamEvent struct {
	TrajectoryItem TrajectoryItem
	Error          error
}

type Handoff struct{}
type DontHandoff struct{}
type Render struct{}
type DontRender struct{}
type ItemIsMessage struct{}
type ItemIsNotMessage struct{}

func (h Handoff) ShouldHandoff() bool {
	return true
}

func (d DontHandoff) ShouldHandoff() bool {
	return false
}

func (r Render) ShouldRender() bool {
	return true
}

func (d DontRender) ShouldRender() bool {
	return false
}

func (i ItemIsMessage) IsMessage() bool {
	return true
}

func (i ItemIsNotMessage) IsMessage() bool {
	return false
}
