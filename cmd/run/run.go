package main

import (
	"context"
	"fmt"
	"webbot/runner"
)

func main() {
	ctx := context.Background()
	runner, err := runner.NewFiniteRunnerFromInitialPageAndRequest(ctx, "https://www.google.com", "search for cats", &runner.RunnerOptions{
		MaxNumSteps: 5,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create runner: %w", err))
	}
	err = runner.Run()
	if err != nil {
		panic(fmt.Errorf("failed to run runner: %w", err))
	}
}
