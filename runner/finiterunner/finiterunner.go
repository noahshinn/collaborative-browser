package finiterunner

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"webbot/actor"
	"webbot/actor/actorstrategy"
	"webbot/browser"
	"webbot/llm"
	"webbot/runner"
	"webbot/trajectory"
	"webbot/utils/io"

	"github.com/yosssi/gohtml"
)

type FiniteRunner struct {
	ctx         context.Context
	actor       actorstrategy.ActorStrategy
	browser     *browser.Browser
	maxNumSteps int
	trajectory  *trajectory.Trajectory
	logPath     string
}

const DefaultMaxNumSteps = 5

type Options struct {
	MaxNumSteps     int
	BrowserOptions  []browser.BrowserOption
	LogPath         string
	ActorStrategyID actor.ActorStrategyID

	BaseActorStrategyID actor.ActorStrategyID
}

const DefaultLogPath = "log"

func NewFiniteRunnerFromInitialPage(ctx context.Context, url string, apiKeys map[string]string, options *Options) (runner.Runner, error) {
	maxNumSteps := DefaultMaxNumSteps
	logPath := DefaultLogPath
	browserOptions := []browser.BrowserOption{}
	actorStrategyID := actor.DefaultActorStrategyID
	baseActorStrategy := actor.DefaultActorStrategyID
	if options != nil {
		if options.MaxNumSteps > 0 {
			maxNumSteps = options.MaxNumSteps
		}
		if options.LogPath != "" {
			logPath = options.LogPath
		}
		if options.BrowserOptions != nil {
			browserOptions = options.BrowserOptions
		}
		if options.ActorStrategyID != "" {
			actorStrategyID = options.ActorStrategyID
		}
		if options.BaseActorStrategyID != "" {
			baseActorStrategy = options.BaseActorStrategyID
		}
	}
	if apiKeys == nil {
		return nil, fmt.Errorf("api keys must be provided")
	} else if openaiApiKey, ok := apiKeys["OPENAI_API_KEY"]; !ok {
		return nil, fmt.Errorf("api keys must contain OPENAI_API_KEY") // for now
	} else {
		userMessage := trajectory.NewUserMessage(fmt.Sprintf("Please go to %s", url))
		allModels := llm.AllModels(openaiApiKey)
		actor, err := actor.ActorStrategyByIDWithOptions(actorStrategyID, allModels, &actor.Options{
			BaseActorStrategyID: baseActorStrategy,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to initialize actor: %w", err)
		}
		browser := browser.NewBrowser(ctx, browserOptions...)
		initialAction := trajectory.NewBrowserNavigateAction(url)
		observation, err := browser.AcceptAction(initialAction.(*trajectory.BrowserAction))
		if err != nil {
			return nil, fmt.Errorf("browser failed to accept initial action: %w", err)
		}
		initialObservation := trajectory.NewBrowserObservation(observation)
		trajectory := &trajectory.Trajectory{
			Items: []trajectory.TrajectoryItem{
				userMessage,
				initialAction,
				initialObservation,
			},
		}
		fmt.Printf("Initializing a finite runner with the following configuration:\n    max num steps: %d\n    actor strategy: %s\n    log path: %s", maxNumSteps, actorStrategyID, logPath)
		return &FiniteRunner{
			ctx:         ctx,
			actor:       actor,
			browser:     browser,
			maxNumSteps: maxNumSteps,
			trajectory:  trajectory,
			logPath:     logPath,
		}, nil
	}
}

func (r *FiniteRunner) Run() error {
	for i := 0; i < r.maxNumSteps; i++ {
		nextAction, err := r.runStep()
		if err != nil {
			return err
		}
		r.trajectory.AddItem(nextAction)
		if nextAction.ShouldHandoff() {
			return nil
		}
		if observation, err := r.browser.AcceptAction(nextAction.(*trajectory.BrowserAction)); err != nil {
			return err
		} else {
			browserDisplay := r.browser.GetDisplay()
			r.trajectory.AddItem(trajectory.NewDebugRenderedDisplay(trajectory.DebugDisplayTypeBrowser, browserDisplay.MD))
			r.trajectory.AddItem(trajectory.NewBrowserObservation(observation))
		}
	}
	r.trajectory.AddItem(trajectory.NewErrorMaxNumStepsReached(r.maxNumSteps))
	return nil
}

func (r *FiniteRunner) runStep() (nextAction trajectory.TrajectoryItem, err error) {
	return r.actor.NextAction(r.ctx, r.trajectory, r.browser)
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
			nextAction, err := r.runStep()
			if err != nil {
				sendErrorTrajectoryItem(err)
				return
			}
			addAndSendTrajectoryItem(nextAction)
			if nextAction.ShouldHandoff() {
				return
			}
			if observation, err := r.browser.AcceptAction(nextAction.(*trajectory.BrowserAction)); err != nil {
				sendErrorTrajectoryItem(err)
				return
			} else {
				browserDisplay := r.browser.GetDisplay()
				addAndSendTrajectoryItem(trajectory.NewDebugRenderedDisplay(trajectory.DebugDisplayTypeBrowser, browserDisplay.MD))
				addAndSendTrajectoryItem(trajectory.NewBrowserObservation(observation))
			}
		}
		addAndSendTrajectoryItem(trajectory.NewErrorMaxNumStepsReached(r.maxNumSteps))
	}()
	return stream, nil
}

func (r *FiniteRunner) Trajectory() *trajectory.Trajectory {
	return r.trajectory
}

func (r *FiniteRunner) Actor() actorstrategy.ActorStrategy {
	return r.actor
}

func (r *FiniteRunner) Browser() *browser.Browser {
	return r.browser
}

func (r *FiniteRunner) Log() error {
	if _, err := os.Stat(r.logPath); os.IsNotExist(err) {
		if err := os.MkdirAll(r.logPath, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}
	trajectoryTextItems := make([]string, len(r.trajectory.Items))
	for i, item := range r.trajectory.Items {
		trajectoryTextItems[i] = item.GetAbbreviatedText()
	}
	browserDisplay := r.browser.GetDisplay()
	if err := io.WriteStringToFile(path.Join(r.logPath, "traj.txt"), strings.Join(trajectoryTextItems, "\n")); err != nil {
		return fmt.Errorf("failed to write trajectory text to file: %w", err)
	} else if err := io.WriteStringToFile(path.Join(r.logPath, "display.md"), browserDisplay.MD); err != nil {
		return fmt.Errorf("failed to write display markdown to file: %w", err)
	} else if err := io.WriteStringToFile(path.Join(r.logPath, "display.html"), gohtml.Format(browserDisplay.HTML)); err != nil {
		return fmt.Errorf("failed to write display html to file: %w", err)
	} else {
		return nil
	}
}
