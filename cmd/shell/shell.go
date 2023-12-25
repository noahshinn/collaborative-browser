package main

import (
	"bufio"
	"collaborativebrowser/actor"
	"collaborativebrowser/afforder"
	"collaborativebrowser/browser"
	"collaborativebrowser/runner/finiterunner"
	"collaborativebrowser/trajectory"
	"collaborativebrowser/utils/printx"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	ctx := context.Background()
	runHeadful := flag.Bool("headful", false, "run the browser in non-headless mode")
	logPath := flag.String("log-path", "out", "the path to write the trajectory and browser display to")
	initialURL := flag.String("url", "https://www.google.com", "the initial url to visit")
	actorStrategy := flag.String("actor-strategy", "base", "the actor strategy to use; one of [\"base\", \"reflexion\"]")
	afforderStrategy := flag.String("afforder-strategy", "function", "the afforder strategy to use; one of [\"function\", filter\"]")
	localStorageServerPort := flag.Int("local-storage-server-port", 2334, "the port to run the local storage server on")
	verbose := flag.Bool("verbose", false, "whether to print verbose debug logs")
	flag.Parse()

	if !*verbose {
		log.SetFlags(0)
		log.SetOutput(io.Discard)
	}

	browserOptions := &browser.Options{
		AttemptToDisableAutomationMessage: true,
	}
	if *runHeadful {
		browserOptions.RunHeadful = true
	}
	if *localStorageServerPort > 0 {
		browserOptions.LocalStorageServerPort = *localStorageServerPort
	}

	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		panic(fmt.Errorf("OPENAI_API_KEY must be set"))
	}
	apiKeys := map[string]string{
		"OPENAI_API_KEY": openaiAPIKey,
	}
	if *actorStrategy == "" {
		*actorStrategy = string(actor.DefaultActorStrategyID)
	}
	if *afforderStrategy == "" {
		*afforderStrategy = string(afforder.DefaultAfforderStrategyID)
	}
	runner, err := finiterunner.NewFiniteRunnerFromInitialPage(ctx, *initialURL, apiKeys, &finiterunner.Options{
		MaxNumSteps:        5,
		BrowserOptions:     browserOptions,
		LogPath:            *logPath,
		ActorStrategyID:    actor.ActorStrategyID(*actorStrategy),
		AfforderStrategyID: afforder.AfforderStrategyID(*afforderStrategy),
	})
	if err != nil {
		panic(fmt.Errorf("failed to create runner: %w", err))
	}
	runner.Log()
	printx.PrintStandardHeader("TRAJECTORY")
	fmt.Println()
	runner.DisplayTrajectory()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("user: ")
ScannerLoop:
	for scanner.Scan() {
		userMessageText := scanner.Text()
		switch userMessageText {
		case "":
			fmt.Print("user: ")
			continue ScannerLoop
		case "headful":
			printx.PrintInColor(printx.ColorGray, "Prepping browser to run in headful mode...")
			if err := runner.RunHeadful(); err != nil {
				printx.PrintInColor(printx.ColorYellow, fmt.Sprintf("Failed to run browser in headful mode: %s, continuing in headless mode", err.Error()))
			}
			printx.PrintInColor(printx.ColorGray, "Running browser in headful mode.")
			fmt.Print("user: ")
			continue ScannerLoop
		case "headless":
			printx.PrintInColor(printx.ColorGray, "Prepping browser to run in headless mode...")
			if err := runner.RunHeadless(); err != nil {
				printx.PrintInColor(printx.ColorYellow, fmt.Sprintf("Failed to run browser in headless mode: %s, continuing in headful mode", err.Error()))
			}
			printx.PrintInColor(printx.ColorGray, "Running browser in headless mode.")
			fmt.Print("user: ")
			continue ScannerLoop
		case "exit":
			fmt.Println("\nexiting...")
			runner.Terminate()
			break ScannerLoop
		case "log":
			runner.Log()
			printx.PrintInColor(printx.ColorGray, "Logged the current state to "+*logPath+".")
			fmt.Print("user: ")
			continue ScannerLoop
		case "help":
			printx.PrintInColor(printx.ColorGray, "This interface is simple - just type natural language. For example, to navigate to google, type \"go to google.com\".\nTo log the current state, type \"log\".\nTo exit gracefully, type \"exit\".")
			fmt.Print("user: ")
			continue ScannerLoop
		default:
			runner.AddItemToTrajectory(trajectory.NewMessage(trajectory.MessageAuthorUser, userMessageText))
			stream, err := runner.RunAndStream()
			if err != nil {
				panic(fmt.Errorf("failed to run and stream: %w", err))
			}
			for event := range stream {
				if event.Error != nil {
					panic(fmt.Errorf("error in stream: %w", event.Error))
				}
				if event.TrajectoryItem.ShouldRender {
					if event.TrajectoryItem.Type == trajectory.TrajectoryItemMessage {
						fmt.Println(event.TrajectoryItem.GetText())
					} else {
						fmt.Println(event.TrajectoryItem.GetAbbreviatedText())
					}
				}
				runner.Log()
			}
			fmt.Print("user: ")
		}
	}
}
