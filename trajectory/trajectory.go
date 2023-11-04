package trajectory

import (
	"fmt"
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
	for _, item := range t.Items[:len(t.Items)-1] {
		if item.ShouldRender() {
			itemTexts = append(itemTexts, item.GetAbbreviatedText())
		}
	}
	itemTexts = append(itemTexts, t.Items[len(t.Items)-1].GetText())
	return fmt.Sprintf("# Trajectory:\n%s", strings.Join(itemTexts, "\n"))
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
