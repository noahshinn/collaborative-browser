package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"webbot/browser"
	"webbot/runner"
	"webbot/runner/finiterunner"
	"webbot/trajectory"
)

func main() {
	ctx := context.Background()
	notHeadless := flag.Bool("not-headless", false, "run the browser in non-headless mode")
	outputFilepath := flag.String("output", "out.txt", "the filepath to write the trajectory to")
	flag.Parse()

	browserOptions := []browser.BrowserOption{
		browser.BrowserOptionAttemptToDisableAutomationMessage,
	}
	if *notHeadless {
		browserOptions = append(browserOptions, browser.BrowserOptionNotHeadless)
	}

	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		panic(fmt.Errorf("OPENAI_API_KEY must be set"))
	}
	runner, err := finiterunner.NewFiniteRunnerFromInitialPage(ctx, "https://www.google.com", &finiterunner.Options{
		MaxNumSteps: 5,
		ApiKeys: map[string]string{
			"OPENAI_API_KEY": openaiAPIKey,
		},
		BrowserOptions: browserOptions,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create runner: %w", err))
	}
	fmt.Println("\n" + strings.Repeat("-", 80) + "\nTRAJECTORY\n" + strings.Repeat("-", 80) + "\n")
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
		}
		runner.Trajectory().AddItem(trajectory.NewUserMessage(userMessageText))
		stream, err := runner.RunAndStream()
		if err != nil {
			writeThenPanic(runner, *outputFilepath, fmt.Errorf("failed to run and stream: %w", err))
		}
		for event := range stream {
			if event.Error != nil {
				writeThenPanic(runner, *outputFilepath, fmt.Errorf("error in stream: %w", event.Error))
				break
			}
			if event.TrajectoryItem.ShouldRender() {
				fmt.Println(event.TrajectoryItem.GetText())
			}
		}
		fmt.Print("user: ")
	}
}

func writeThenPanic(r runner.Runner, filepath string, er error) {
	if err := r.Log(filepath); err != nil {
		panic(fmt.Errorf("failed to log trajectory: %w", er))
	} else {
		panic(er)
	}
}
