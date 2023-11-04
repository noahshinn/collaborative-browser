package finiterunner

import (
	"context"
	"fmt"
	"strings"
	act "webbot/actor"
	"webbot/browser"
	"webbot/browser/language"
	"webbot/llm"
	"webbot/runner"
	"webbot/trajectory"
	"webbot/utils/io"
)

type FiniteRunner struct {
	ctx         context.Context
	actor       act.Actor
	browser     *browser.Browser
	maxNumSteps int
	trajectory  *trajectory.Trajectory
}

const DefaultMaxNumSteps = 5

type Options struct {
	MaxNumSteps    int
	BrowserOptions []browser.BrowserOption
	ApiKeys        map[string]string
}

func NewFiniteRunnerFromInitialPage(ctx context.Context, url string, options *Options) (runner.Runner, error) {
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
			actor:       actor,
			browser:     browser,
			maxNumSteps: maxNumSteps,
			trajectory:  trajectory,
		}, nil
	}
}

func NewFiniteRunnerFromInitialPageAndRequest(ctx context.Context, url string, request string, options *Options) (runner.Runner, error) {
	runner, err := NewFiniteRunnerFromInitialPage(ctx, url, options)
	if err != nil {
		return nil, err
	}
	message := trajectory.NewUserMessage(request)
	runner.Trajectory().AddItem(message)
	nextAction, debugMessageDisplay, err := runner.Actor().NextAction(ctx, runner.Trajectory().GetText())
	if err != nil {
		return nil, fmt.Errorf("page visit was successful but the actor failed to perform the initial action: %w", err)
	}
	runner.Trajectory().AddItem(debugMessageDisplay)
	runner.Trajectory().AddItem(nextAction)
	if nextAction.ShouldHandoff() {
		return runner, nil
	} else {
		if err := runner.Browser().AcceptAction(nextAction.(*trajectory.BrowserAction)); err != nil {
			return nil, fmt.Errorf("page visit was successful but the browser failed to accept the initial action: %w", err)
		} else if location, pageRender, err := runner.Browser().Render(language.LanguageMD); err != nil {
			return nil, fmt.Errorf("page visit was successful but the browser failed to render the initial page: %w", err)
		} else {
			runner.Trajectory().AddItem(trajectory.NewBrowserObservation(pageRender, location))
			return runner, nil
		}
	}
}

func (r *FiniteRunner) Run() error {
	for i := 0; i < r.maxNumSteps; i++ {
		nextAction, debugDisplays, err := r.runStep()
		r.trajectory.AddItems(debugDisplays)
		if err != nil {
			return err
		}
		r.trajectory.AddItem(nextAction)
		if !nextAction.ShouldHandoff() {
			// // TODO: store the previous page render so that it doesn't have to be rerendered
			location, pageRender, err := r.browser.Render(language.LanguageMD)
			if err != nil {
				return fmt.Errorf("browser failed to render page: %w", err)
			}
			r.trajectory.AddItem(trajectory.NewBrowserObservation(pageRender, location))
		}
	}
	r.trajectory.AddItem(trajectory.NewErrorMaxNumStepsReached(r.maxNumSteps))
	return nil
}

func (r *FiniteRunner) runStep() (nextAction trajectory.TrajectoryItem, debugDisplays []trajectory.TrajectoryItem, err error) {
	state := r.trajectory.GetText()
	nextAction, debugMessageDisplay, err := r.actor.NextAction(r.ctx, state)
	if debugMessageDisplay == nil {
		return nextAction, []trajectory.TrajectoryItem{}, nil
	}
	return nextAction, []trajectory.TrajectoryItem{debugMessageDisplay}, err
}

func (r *FiniteRunner) RunAndStream() (<-chan *trajectory.TrajectoryStreamEvent, error) {
	stream := make(chan *trajectory.TrajectoryStreamEvent)
	addAndSendTrajectoryItem := func(item trajectory.TrajectoryItem) {
		r.trajectory.AddItem(item)
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
		for i := 0; i < r.maxNumSteps; i++ {
			nextAction, debugDisplays, err := r.runStep()
			if err != nil {
				sendErrorTrajectoryItem(err)
				return
			}
			for _, debugDisplay := range debugDisplays {
				if debugDisplay != nil {
					addAndSendTrajectoryItem(debugDisplay)
				}
			}
			addAndSendTrajectoryItem(nextAction)
			if nextAction.ShouldHandoff() {
				return
			}
			// TODO: store the previous page render so that it doesn't have to be rerendered
			location, pageRender, err := r.browser.Render(language.LanguageMD)
			if err != nil {
				sendErrorTrajectoryItem(fmt.Errorf("browser failed to render page: %w", err))
				return
			}
			addAndSendTrajectoryItem(trajectory.NewDebugRenderedDisplay(trajectory.DebugDisplayTypeBrowser, pageRender))
			addAndSendTrajectoryItem(trajectory.NewBrowserObservation(pageRender, location))
		}
		addAndSendTrajectoryItem(trajectory.NewErrorMaxNumStepsReached(r.maxNumSteps))
	}()
	return stream, nil
}

func (r *FiniteRunner) Trajectory() *trajectory.Trajectory {
	return r.trajectory
}

func (r *FiniteRunner) Actor() act.Actor {
	return r.actor
}

func (r *FiniteRunner) Browser() *browser.Browser {
	return r.browser
}

func (r *FiniteRunner) Log(filepath string) error {
	trajectoryTextItems := make([]string, len(r.trajectory.Items))
	for i, item := range r.trajectory.Items {
		trajectoryTextItems[i] = item.GetText()
	}
	if err := io.WriteStringToFile(filepath, strings.Join(trajectoryTextItems, "\n")); err != nil {
		return fmt.Errorf("failed to write trajectory text to file: %w", err)
	}
	return nil
}
