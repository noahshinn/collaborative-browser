package main

import (
	"bufio"
	"collaborativebrowser/actor"
	"collaborativebrowser/browser"
	"collaborativebrowser/runner/finiterunner"
	"collaborativebrowser/trajectory"
	"collaborativebrowser/utils/printx"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	ctx := context.Background()
	runHeadful := flag.Bool("headful", false, "run the browser in non-headless mode")
	logPath := flag.String("log-path", "out", "the path to write the trajectory and browser display to")
	initialURL := flag.String("url", "https://www.google.com", "the initial url to visit")
	actorStrategy := flag.String("actor-strategy", "base_llm", "the actor strategy to use; one of [base_llm, reflexion]")
	flag.Parse()

	browserOptions := []browser.BrowserOption{
		browser.BrowserOptionAttemptToDisableAutomationMessage,
	}
	if *runHeadful {
		browserOptions = append(browserOptions, browser.BrowserOptionHeadful)
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
	runner, err := finiterunner.NewFiniteRunnerFromInitialPage(ctx, *initialURL, apiKeys, &finiterunner.Options{
		MaxNumSteps:     5,
		BrowserOptions:  browserOptions,
		LogPath:         *logPath,
		ActorStrategyID: actor.ActorStrategyID(*actorStrategy),
	})
	if err != nil {
		panic(fmt.Errorf("failed to create runner: %w", err))
	}
	runner.Log()
	printx.PrintStandardHeader("TRAJECTORY")
	fmt.Println()
	for _, item := range runner.Trajectory().Items {
		fmt.Println(item.GetAbbreviatedText())
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("user: ")
	for scanner.Scan() {
		userMessageText := scanner.Text()
		if len(userMessageText) == 0 {
			fmt.Print("user: ")
			continue
		} else if userMessageText == "exit" {
			fmt.Println("\nexiting...")
			break
		} else {
			userMessageText = strings.TrimSpace(userMessageText)
		}
		runner.Trajectory().AddItem(trajectory.NewMessage(trajectory.MessageAuthorUser, userMessageText))
		stream, err := runner.RunAndStream()
		if err != nil {
			panic(fmt.Errorf("failed to run and stream: %w", err))
		}
		for event := range stream {
			if event.Error != nil {
				panic(fmt.Errorf("error in stream: %w", event.Error))
			}
			if event.TrajectoryItem.ShouldRender() {
				if _, ok := event.TrajectoryItem.(*trajectory.Message); ok {
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
