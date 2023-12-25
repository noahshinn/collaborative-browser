package trajectory

func NewErrorMaxNumStepsReached(maxNumSteps int) *TrajectoryItem {
	return &TrajectoryItem{
		Type:          TrajectoryItemMaxNumStepsReached,
		ShouldHandoff: true,
		ShouldRender:  true,
		MaxNumSteps:   maxNumSteps,
	}
}

func NewErrorMaxContextLengthExceeded(contextLengthAllowed int, contextLengthReceived int) *TrajectoryItem {
	return &TrajectoryItem{
		Type:                  TrajectoryItemMaxContextLengthExceeded,
		ShouldHandoff:         true,
		ShouldRender:          true,
		ContextLengthAllowed:  contextLengthAllowed,
		ContextLengthReceived: contextLengthReceived,
	}
}
