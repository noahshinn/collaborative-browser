package runner

import (
	"context"
	"fmt"
	act "webbot/actor"
	"webbot/browser"
	"webbot/browser/language"
	"webbot/llm"
	"webbot/trajectory"
)

type Runner interface {
	Run() error
	RunAndStream() (<-chan trajectory.TrajectoryStreamEvent, error)
}

type FiniteRunner struct {
	ctx         context.Context
	Actor       act.Actor
	Browser     *browser.Browser
	MaxNumSteps int
	Trajectory  *trajectory.Trajectory
}

const DefaultMaxNumSteps = 5

type RunnerOptions struct {
	MaxNumSteps    int
	BrowserOptions []browser.BrowserOption
	ApiKeys        map[string]string
}

func NewFiniteRunnerFromInitialPage(ctx context.Context, url string, options *RunnerOptions) (*FiniteRunner, error) {
	maxNumSteps := DefaultMaxNumSteps
	if options.MaxNumSteps > 0 {
		maxNumSteps = options.MaxNumSteps
	}
	if options.ApiKeys == nil {
		return nil, fmt.Errorf("api keys must be provided")
	} else if openaiApiKey, ok := options.ApiKeys["OPENAI_API_KEY"]; !ok {
		return nil, fmt.Errorf("api keys must contain OPENAI_API_KEY") // for now
	} else {
		userMessage := trajectory.NewUserMessage(fmt.Sprintf("Please go to %s", url))
		allModels := llm.AllModels(openaiApiKey)
		actor := act.NewLLMActor(allModels.DefaultChatModel)
		browser := browser.NewBrowser(ctx, options.BrowserOptions...)
		initialAction := trajectory.NewBrowserNavigateAction(url)
		if err := browser.AcceptAction(initialAction.(*trajectory.BrowserAction)); err != nil {
			return nil, fmt.Errorf("browser failed to accept initial action: %w", err)
		}
		location, pageRender, err := browser.Render(language.LanguageMD)
		if err != nil {
			return nil, fmt.Errorf("page visit was successful but browser failed to render initial page: %w", err)
		}
		initialObservation := trajectory.NewBrowserObservation(pageRender, location)
		trajectory := &trajectory.Trajectory{
			Items: []trajectory.TrajectoryItem{
				userMessage,
				initialAction,
				initialObservation,
			},
		}
		return &FiniteRunner{
			ctx:         ctx,
			Actor:       actor,
			Browser:     browser,
			MaxNumSteps: maxNumSteps,
			Trajectory:  trajectory,
		}, nil
	}
}

func NewFiniteRunnerFromInitialPageAndRequest(ctx context.Context, url string, request string, options *RunnerOptions) (*FiniteRunner, error) {
	runner, err := NewFiniteRunnerFromInitialPage(ctx, url, options)
	if err != nil {
		return nil, err
	}
	message := trajectory.NewUserMessage(request)
	runner.Trajectory.AddItem(message)
	nextAction, debugMessageDisplay, err := runner.Actor.NextAction(ctx, runner.Trajectory.GetText())
	if err != nil {
		return nil, fmt.Errorf("page visit was successful but the actor failed to perform the initial action: %w", err)
	}
	runner.Trajectory.AddItem(debugMessageDisplay)
	runner.Trajectory.AddItem(nextAction)
	if nextAction.ShouldHandoff() {
		return runner, nil
	} else {
		if err := runner.Browser.AcceptAction(nextAction.(*trajectory.BrowserAction)); err != nil {
			return nil, fmt.Errorf("page visit was successful but the browser failed to accept the initial action: %w", err)
		} else if location, pageRender, err := runner.Browser.Render(language.LanguageMD); err != nil {
			return nil, fmt.Errorf("page visit was successful but the browser failed to render the initial page: %w", err)
		} else {
			runner.Trajectory.AddItem(trajectory.NewBrowserObservation(pageRender, location))
			return runner, nil
		}
	}
}

func (r *FiniteRunner) Run() error {
	for i := 0; i < r.MaxNumSteps; i++ {
		nextAction, debugDisplays, err := r.runStep()
		r.Trajectory.AddItems(debugDisplays)
		if err != nil {
			return err
		}
		r.Trajectory.AddItem(nextAction)
		if !nextAction.ShouldHandoff() {
			// // TODO: store the previous page render so that it doesn't have to be rerendered
			location, pageRender, err := r.Browser.Render(language.LanguageMD)
			if err != nil {
				return fmt.Errorf("browser failed to render page: %w", err)
			}
			r.Trajectory.AddItem(trajectory.NewBrowserObservation(pageRender, location))
		}
	}
	r.Trajectory.AddItem(trajectory.NewErrorMaxNumStepsReached(r.MaxNumSteps))
	return nil
}

func (r *FiniteRunner) runStep() (nextAction trajectory.TrajectoryItem, debugDisplays []trajectory.TrajectoryItem, err error) {
	state := r.Trajectory.GetText()
	nextAction, debugMessageDisplay, err := r.Actor.NextAction(r.ctx, state)
	return nextAction, []trajectory.TrajectoryItem{debugMessageDisplay}, err
}

func (r *FiniteRunner) RunAndStream() (<-chan *trajectory.TrajectoryStreamEvent, error) {
	stream := make(chan *trajectory.TrajectoryStreamEvent)
	sendTrajectoryItem := func(item trajectory.TrajectoryItem) {
		stream <- &trajectory.TrajectoryStreamEvent{
			TrajectoryItem: item,
		}
	}
	sendErrorTrajectoryItem := func(err error) {
		stream <- &trajectory.TrajectoryStreamEvent{
			Error: err,
		}
	}
	go func() {
		defer close(stream)
		for i := 0; i < r.MaxNumSteps; i++ {
			nextAction, debugDisplays, err := r.runStep()
			r.Trajectory.AddItems(debugDisplays)
			if err != nil {
				sendErrorTrajectoryItem(err)
				return
			}
			r.Trajectory.AddItem(nextAction)
			sendTrajectoryItem(nextAction)
			if nextAction.ShouldHandoff() {
				return
			}
			// TODO: store the previous page render so that it doesn't have to be rerendered
			location, pageRender, err := r.Browser.Render(language.LanguageMD)
			if err != nil {
				stream <- &trajectory.TrajectoryStreamEvent{
					Error: fmt.Errorf("browser failed to render page: %w", err),
				}
				return
			}
			debugBrowserDisplay := trajectory.NewDebugRenderedDisplay(trajectory.DebugDisplayTypeBrowser, pageRender)
			observation := trajectory.NewBrowserObservation(pageRender, location)
			r.Trajectory.AddItem(debugBrowserDisplay)
			r.Trajectory.AddItem(observation)
			sendTrajectoryItem(debugBrowserDisplay)
			sendTrajectoryItem(observation)
		}
		errorMaxNumStepsReached := trajectory.NewErrorMaxNumStepsReached(r.MaxNumSteps)
		r.Trajectory.AddItem(errorMaxNumStepsReached)
		sendTrajectoryItem(errorMaxNumStepsReached)
	}()
	return stream, nil
}
