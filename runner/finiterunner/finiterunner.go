package finiterunner

import (
	"collaborativebrowser/actor"
	"collaborativebrowser/actor/actorstrategy"
	"collaborativebrowser/afforder"
	"collaborativebrowser/browser"
	"collaborativebrowser/llm"
	"collaborativebrowser/runner"
	"collaborativebrowser/runner/sharedlocalstorage"
	"collaborativebrowser/trajectory"
	"collaborativebrowser/utils/io"
	"collaborativebrowser/utils/printx"
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/yosssi/gohtml"
)

type FiniteRunner struct {
	ctx                context.Context
	actor              actorstrategy.ActorStrategy
	browser            *browser.Browser
	maxNumSteps        int
	logPath            string
	sharedLocalStorage *sharedlocalstorage.SharedLocalStorage
}

const DefaultMaxNumSteps = 5

type Options struct {
	MaxNumSteps        int
	BrowserOptions     *browser.Options
	LogPath            string
	ActorStrategyID    actor.ActorStrategyID
	AfforderStrategyID afforder.AfforderStrategyID
}

const DefaultLogPath = "log"

func NewFiniteRunnerFromInitialPage(ctx context.Context, url string, apiKeys map[string]string, options *Options) (runner.Runner, error) {
	maxNumSteps := DefaultMaxNumSteps
	logPath := DefaultLogPath
	browserOptions := &browser.Options{}
	actorStrategyID := actor.DefaultActorStrategyID
	afforderStrategyID := afforder.DefaultAfforderStrategyID
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
		if options.AfforderStrategyID != "" {
			afforderStrategyID = options.AfforderStrategyID
		}
	}
	if apiKeys == nil {
		return nil, fmt.Errorf("api keys must be provided")
	} else if openaiApiKey, ok := apiKeys["OPENAI_API_KEY"]; !ok {
		return nil, fmt.Errorf("api keys must contain OPENAI_API_KEY") // for now
	} else {
		b := browser.NewBrowser(ctx, browserOptions)
		sls := sharedlocalstorage.NewSharedLocalStorage(ctx, nil)
		if err := sls.Start(); err != nil {
			return nil, fmt.Errorf("failed to start shared local storage: %w", err)
		}
		userMessage := trajectory.NewMessage(trajectory.MessageAuthorUser, fmt.Sprintf("Please go to %s", url))
		allModels := llm.AllModels(openaiApiKey)
		actor, err := actor.ActorStrategyByIDWithOptions(actorStrategyID, allModels, &actor.Options{
			AfforderStrategyID: afforderStrategyID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to initialize actor: %w", err)
		}
		initialAction := trajectory.NewBrowserNavigateAction(url)
		observation, err := b.AcceptAction(initialAction)
		if err != nil {
			return nil, fmt.Errorf("browser failed to accept initial action: %w", err)
		}
		initialObservation := trajectory.NewBrowserObservation(observation)
		r := &FiniteRunner{
			ctx:                ctx,
			actor:              actor,
			browser:            b,
			maxNumSteps:        maxNumSteps,
			logPath:            logPath,
			sharedLocalStorage: sls,
		}
		if err := r.AddItemsToTrajectory([]*trajectory.TrajectoryItem{
			userMessage,
			initialAction,
			initialObservation,
		}); err != nil {
			return nil, fmt.Errorf("failed to add items to shared trajectory: %w", err)
		}
		printx.PrintStandardHeader("CONFIGURATION")
		fmt.Printf("\nInitializing a finite runner with the following configuration:\n- Maximum number steps per turn: %d\n- Actor strategy: %s\n- Log path: %s\n", maxNumSteps, actorStrategyID, logPath)
		return r, nil
	}
}

func (r *FiniteRunner) Run() error {
	for i := 0; i < r.maxNumSteps; i++ {
		nextAction, err := r.runStep()
		if err != nil {
			return err
		}
		if err := r.AddItemToTrajectory(nextAction); err != nil {
			return err
		}
		if nextAction.ShouldHandoff {
			return nil
		}
		if observation, err := r.browser.AcceptAction(nextAction); err != nil {
			return err
		} else if err := r.AddItemToTrajectory(trajectory.NewBrowserObservation(observation)); err != nil {
			return err
		}
	}
	return r.AddItemToTrajectory(trajectory.NewErrorMaxNumStepsReached(r.maxNumSteps))
}

func (r *FiniteRunner) runStep() (nextAction *trajectory.TrajectoryItem, err error) {
	traj, err := r.readTrajectory()
	if err != nil {
		return nil, err
	}
	return r.actor.NextAction(r.ctx, traj, r.browser)
}

func (r *FiniteRunner) RunAndStream() (<-chan *trajectory.TrajectoryStreamEvent, error) {
	stream := make(chan *trajectory.TrajectoryStreamEvent)
	addAndSendTrajectoryItem := func(item *trajectory.TrajectoryItem) error {
		if err := r.AddItemToTrajectory(item); err != nil {
			return err
		}
		stream <- &trajectory.TrajectoryStreamEvent{
			TrajectoryItem: item,
		}
		return nil
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
			if nextAction.ShouldHandoff {
				return
			}
			if observation, err := r.browser.AcceptAction(nextAction); err != nil {
				sendErrorTrajectoryItem(err)
				return
			} else if err := addAndSendTrajectoryItem(trajectory.NewBrowserObservation(observation)); err != nil {
				sendErrorTrajectoryItem(err)
				return
			}
		}
		if err := addAndSendTrajectoryItem(trajectory.NewErrorMaxNumStepsReached(r.maxNumSteps)); err != nil {
			sendErrorTrajectoryItem(err)
			return
		}
	}()
	return stream, nil
}

func (r *FiniteRunner) DisplayTrajectory() error {
	traj, err := r.readTrajectory()
	if err != nil {
		return err
	}
	for _, item := range traj.Items {
		fmt.Println(item.GetAbbreviatedText())
	}
	return nil
}

func (r *FiniteRunner) readTrajectory() (*trajectory.Trajectory, error) {
	return r.sharedLocalStorage.ReadTrajectory()
}

func (r *FiniteRunner) AddItemToTrajectory(item *trajectory.TrajectoryItem) error {
	return r.AddItemsToTrajectory([]*trajectory.TrajectoryItem{item})
}

func (r *FiniteRunner) AddItemsToTrajectory(items []*trajectory.TrajectoryItem) error {
	return r.sharedLocalStorage.AddItemsToTrajectory(items)
}

func (r *FiniteRunner) Log() error {
	if _, err := os.Stat(r.logPath); os.IsNotExist(err) {
		if err := os.MkdirAll(r.logPath, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}
	traj, err := r.readTrajectory()
	if err != nil {
		return err
	}
	trajectoryTextItems := make([]string, len(traj.Items))
	for i, item := range traj.Items {
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

func (r *FiniteRunner) RunHeadful() error {
	return r.browser.RunHeadful(r.ctx)
}

func (r *FiniteRunner) RunHeadless() error {
	return r.browser.RunHeadless(r.ctx)
}

func (r *FiniteRunner) Terminate() {
	r.browser.Cancel()
	r.sharedLocalStorage.Stop()
}
