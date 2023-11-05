package trajectory

import "fmt"

type errorItem struct {
	Handoff
	Render
}

type ErrorMaxNumStepsReached struct {
	errorItem
	MaxNumSteps int
}

func NewErrorMaxNumStepsReached(maxNumSteps int) TrajectoryItem {
	return &ErrorMaxNumStepsReached{
		MaxNumSteps: maxNumSteps,
	}
}

type ErrorMaxContextLengthExceeded struct {
	errorItem
	ContextLengthAllowed  int
	ContextLengthReceived int
}

func NewErrorMaxContextLengthExceeded(contextLengthAllowed int, contextLengthReceived int) TrajectoryItem {
	return &ErrorMaxContextLengthExceeded{
		ContextLengthAllowed:  contextLengthAllowed,
		ContextLengthReceived: contextLengthReceived,
	}
}

func (m *ErrorMaxNumStepsReached) GetText() string {
	return fmt.Sprintf("error: max num steps reached: %d", m.MaxNumSteps)
}

func (m *ErrorMaxNumStepsReached) GetAbbreviatedText() string {
	return m.GetText()
}

func (m *ErrorMaxContextLengthExceeded) GetText() string {
	return fmt.Sprintf("error: max context length exceeded: %d > %d tokens", m.ContextLengthReceived, m.ContextLengthAllowed)
}

func (m *ErrorMaxContextLengthExceeded) GetAbbreviatedText() string {
	return m.GetText()
}
