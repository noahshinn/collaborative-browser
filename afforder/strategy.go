package afforder

import (
	"collaborativebrowser/afforder/afforderstrategy"
	"collaborativebrowser/afforder/afforderstrategy/functionafforder"
	"log"
)

type AfforderStrategyID string

const (
	AfforderStrategyIDFunctionAfforder AfforderStrategyID = "function_afforder"
)

const DefaultAfforderStrategyID = AfforderStrategyIDFunctionAfforder

func AfforderStrategyByID(id AfforderStrategyID) afforderstrategy.AfforderStrategy {
	switch id {
	case AfforderStrategyIDFunctionAfforder:
		return functionafforder.New()
	default:
		log.Printf("invalid afforder strategy ID: %s; defaulting to %s", id, DefaultAfforderStrategyID)
		return functionafforder.New()
	}
}
