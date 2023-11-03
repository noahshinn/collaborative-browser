package runner

import (
	"context"
	"fmt"
	act "webbot/actor"
	"webbot/browser"
	"webbot/browser/language"
	"webbot/llm"
	"webbot/runner/trajectory"
)

type Runner interface {
	Run() error
}

type FiniteRunner struct {
	Actor       act.Actor
	Browser     *browser.Browser
	MaxNumSteps int
	Trajectory  []trajectory.TrajectoryItem
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
		allModels := llm.AllModels(openaiApiKey)
		actor := act.NewLLMActor(allModels.DefaultChatModel)
		browser := browser.NewBrowser(ctx, options.BrowserOptions...)
		initialAction := trajectory.NewBrowserNavigateAction(url)
		if err := browser.AcceptAction(initialAction); err != nil {
			return nil, fmt.Errorf("browser failed to accept initial action: %w", err)
		}
		pageRender, err := browser.Render(language.LanguageMD)
		if err != nil {
			return nil, fmt.Errorf("page visit was successful but browser failed to render initial page: %w", err)
		}
		initialObservation := trajectory.NewBrowserObservation(pageRender)
		trajectory := []trajectory.TrajectoryItem{initialAction, initialObservation}
		return &FiniteRunner{
			Actor:       actor,
			Browser:     browser,
			MaxNumSteps: maxNumSteps,
			Trajectory:  trajectory,
		}, nil
	}
}

func NewFiniteRunnerFromInitialPageAndRequest(url string, request string, maxNumSteps int, options *RunnerOptions) *FiniteRunner {
	// TODO: implement
	return nil
}

func (r *FiniteRunner) Run() error {
	for i := 0; i < r.MaxNumSteps; i++ {
		if err := r.runStep(); err != nil {
			return err
		}
	}
	return nil
}

func (r *FiniteRunner) runStep() error {
	// TODO: implement
	return nil
}
