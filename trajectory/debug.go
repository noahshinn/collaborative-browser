package trajectory

type DebugDisplayType string

const (
	DebugDisplayTypeBrowser    DebugDisplayType = "browser"
	DebugDisplayTypeTrajectory DebugDisplayType = "trajectory"
	DebugDisplayTypeLLMMesages DebugDisplayType = "llm_messages"
)

type DebugRenderedDisplay struct {
	DontHandoff
	DontRender

	Type DebugDisplayType
	Text string
}

func NewDebugRenderedDisplay(typ DebugDisplayType, text string) TrajectoryItem {
	return &DebugRenderedDisplay{
		Type: typ,
		Text: text,
	}
}

func (d *DebugRenderedDisplay) GetText() string {
	return d.Text
}

func (d *DebugRenderedDisplay) GetAbbreviatedText() string {
	return d.Text
}
