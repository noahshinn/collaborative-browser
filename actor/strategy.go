package actor

import (
	"collaborativebrowser/actor/actorstrategy"
	"collaborativebrowser/actor/actorstrategy/basellm"
	"collaborativebrowser/actor/actorstrategy/react"
	"collaborativebrowser/actor/actorstrategy/reflexion"
	"collaborativebrowser/actor/actorstrategy/verification"
	"collaborativebrowser/afforder"
	"collaborativebrowser/llm"
	"fmt"
	"log"
)

type ActorStrategyID string

const (
	ActorStrategyIDBaseLLM      ActorStrategyID = "base"
	ActorStrategyIDReact        ActorStrategyID = "react"
	ActorStrategyIDVerification ActorStrategyID = "verification"
	ActorStrategyIDReflexion    ActorStrategyID = "reflexion"
)

const DefaultActorStrategyID = ActorStrategyIDBaseLLM

func DefaultActorStrategy(models *llm.Models) actorstrategy.ActorStrategy {
	actorStrategy, err := ActorStrategyByIDWithOptions(DefaultActorStrategyID, models, nil)
	if err != nil {
		log.Printf("error getting default actor strategy, defaulting to `base-llm`: %v", err)
		return basellm.New(models, nil)
	}
	return actorStrategy
}

type Options struct {
	AfforderStrategyID afforder.AfforderStrategyID

	// for reflexion
	MaxNumIterations int
}

func ActorStrategyByIDWithOptions(strategyID ActorStrategyID, models *llm.Models, options *Options) (actorstrategy.ActorStrategy, error) {
	switch strategyID {
	case ActorStrategyIDBaseLLM:
		return basellm.New(models, &actorstrategy.Options{
			AfforderStrategyID: options.AfforderStrategyID,
		}), nil
	case ActorStrategyIDReact:
		return react.New(models, &actorstrategy.Options{
			AfforderStrategyID: options.AfforderStrategyID,
		}), nil
	case ActorStrategyIDVerification:
		return verification.New(models, nil), nil
	case ActorStrategyIDReflexion:
		return reflexion.New(models, &actorstrategy.Options{
			MaxNumIterations: options.MaxNumIterations,
		}), nil
	}
	return nil, fmt.Errorf("invalid actor strategy ID: %s", strategyID)
}
