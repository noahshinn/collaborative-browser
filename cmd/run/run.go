package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"webbot/runner"
	"webbot/runner/trajectory"
)

func main() {
	ctx := context.Background()
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		panic(fmt.Errorf("OPENAI_API_KEY must be set"))
	}
	runner, err := runner.NewFiniteRunnerFromInitialPage(ctx, "https://www.google.com", &runner.RunnerOptions{
		MaxNumSteps: 5,
		ApiKeys: map[string]string{
			"OPENAI_API_KEY": openaiAPIKey,
		},
	})
	if err != nil {
		panic(fmt.Errorf("failed to create runner: %w", err))
	}
	fmt.Println(strings.Repeat("-", 80) + "\n" + "TRAJECTORY" + "\n" + strings.Repeat("-", 80))
	for _, item := range runner.Trajectory.Items {
		fmt.Println(item.GetAbbreviatedText())
	}

	for {
		userInput := getUserInput()
		runner.Trajectory.AddItem(trajectory.NewUserMessage(userInput))
		stream, err := runner.RunAndStream()
		if err != nil {
			panic(fmt.Errorf("failed to run and stream: %w", err))
		}
		for event := range stream {
			if event.Error != nil {
				panic(fmt.Errorf("error in stream: %w", event.Error))
			} else {
				fmt.Println(event.TrajectoryItem.GetText())
			}
		}
	}
}

func getUserInput() string {
	var input string
	fmt.Print("user: ")
	fmt.Scanln(&input)
	return input
}
