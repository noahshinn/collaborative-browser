package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"webbot/runner"
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

	for {
		stream, err := runner.RunAndStream()
		if err != nil {
			panic(fmt.Errorf("failed to run and stream: %w", err))
		}
		fmt.Println(strings.Repeat("-", 80) + "\n" + "TRAJECTORY" + "\n" + strings.Repeat("-", 80))
		for event := range stream {
			if event.Error != nil {
				panic(fmt.Errorf("error in stream: %w", event.Error))
			} else {
				fmt.Println(event.TrajectoryItem.GetText())
			}
		}
	}
}
