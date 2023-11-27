package actor

import (
	"collaborativebrowser/actor/actorstrategy"
	"collaborativebrowser/actor/actorstrategy/basellm"
	"collaborativebrowser/actor/actorstrategy/react"
	"collaborativebrowser/actor/actorstrategy/reflexion"
	"collaborativebrowser/actor/actorstrategy/verification"
	"collaborativebrowser/llm"
	"fmt"
	"log"
)

type ActorStrategyID string

const (
	ActorStrategyIDBaseLLM      ActorStrategyID = "base_llm"
	ActorStrategyIDReact        ActorStrategyID = "react"
	ActorStrategyIDVerification ActorStrategyID = "verification"
	ActorStrategyIDReflexion    ActorStrategyID = "reflexion"
)

const DefaultActorStrategyID = ActorStrategyIDBaseLLM

func DefaultActorStrategy(models *llm.Models) actorstrategy.ActorStrategy {
	actorStrategy, err := ActorStrategyByIDWithOptions(DefaultActorStrategyID, models, nil)
	if err != nil {
		log.Printf("error getting default actor strategy, defaulting to `base-llm`: %v", err)
		return basellm.New(models)
	}
	return actorStrategy
}

type Options struct {
	// for verification and reflexion
	BaseActorStrategyID ActorStrategyID

	// for reflexion
	MaxNumIterations int
}

func ActorStrategyByIDWithOptions(strategyID ActorStrategyID, models *llm.Models, options *Options) (actorstrategy.ActorStrategy, error) {
	switch strategyID {
	case ActorStrategyIDBaseLLM:
		return basellm.New(models), nil
	case ActorStrategyIDReact:
		return react.New(models), nil
	case ActorStrategyIDVerification:
		baseActorStrategyID := DefaultActorStrategyID
		if options != nil {
			if options.BaseActorStrategyID != "" {
				if options.BaseActorStrategyID == strategyID {
					return nil, fmt.Errorf("invalid base actor strategy ID: %s; this will cause an infinite loop", options.BaseActorStrategyID)
				}
				baseActorStrategyID = options.BaseActorStrategyID
			}
		}
		baseActorStrategy, err := ActorStrategyByIDWithOptions(baseActorStrategyID, models, options)
		if err != nil {
			return nil, fmt.Errorf("error getting base actor strategy: %w", err)
		}
		return verification.New(models, baseActorStrategy), nil
	case ActorStrategyIDReflexion:
		baseActorStrategyID := DefaultActorStrategyID
		maxNumIterations := reflexion.DefaultMaxNumIterations
		if options != nil {
			if options.BaseActorStrategyID != "" {
				if options.BaseActorStrategyID == strategyID {
					return nil, fmt.Errorf("invalid base actor strategy ID: %s; this will cause an infinite loop", options.BaseActorStrategyID)
				}
				baseActorStrategyID = options.BaseActorStrategyID
			}
		}
		baseActorStrategy, err := ActorStrategyByIDWithOptions(baseActorStrategyID, models, options)
		if err != nil {
			return nil, fmt.Errorf("error getting base actor strategy: %w", err)
		}
		return reflexion.New(models, baseActorStrategy, maxNumIterations), nil
	}
	return nil, fmt.Errorf("invalid actor strategy ID: %s", strategyID)
}
