package trajectory

import (
	"strings"
	"webbot/utils/slicesx"
)

type Trajectory struct {
	Items []TrajectoryItem
}

func (t *Trajectory) GetText() string {
	if len(t.Items) == 0 {
		return ""
	}
	itemTexts := slicesx.Map(t.Items, func(item TrajectoryItem) string {
		return item.GetAbbreviatedText()
	})
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
}

type TrajectoryStreamEvent struct {
	TrajectoryItem TrajectoryItem
	Error          error
}

type DontHandoff struct{}
type DontRender struct{}

func (d DontRender) ShouldRender() bool {
	return false
}

func (d DontHandoff) ShouldHandoff() bool {
	return false
}
