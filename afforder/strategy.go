package afforder

import (
	"collaborativebrowser/afforder/afforderstrategy"
	"collaborativebrowser/afforder/afforderstrategy/filterafforder"
	"collaborativebrowser/afforder/afforderstrategy/functionafforder"
	"log"
)

type AfforderStrategyID string

const (
	AfforderStrategyIDFunctionAfforder AfforderStrategyID = "function_afforder"
	AfforderStrategyIDFilterAfforder   AfforderStrategyID = "filter_afforder"
)

const DefaultAfforderStrategyID = AfforderStrategyIDFunctionAfforder

func AfforderStrategyByID(id AfforderStrategyID) afforderstrategy.AfforderStrategy {
	switch id {
	case AfforderStrategyIDFunctionAfforder:
		return functionafforder.New()
	case AfforderStrategyIDFilterAfforder:
		return filterafforder.New()
	default:
		log.Printf("invalid afforder strategy ID: %s; defaulting to %s", id, DefaultAfforderStrategyID)
		return functionafforder.New()
	}
}
