package trajectory

import (
	"fmt"
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
	itemTexts := slicesx.Map(t.Items[:len(t.Items)-1], func(item TrajectoryItem) string {
		return item.GetText()
	})
	lastItem := t.Items[len(t.Items)-1].GetText()
	return fmt.Sprintf("# Trajectory:\n%s", strings.Join(append(itemTexts, lastItem), "\n"))
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
