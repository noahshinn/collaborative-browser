package afforder

import (
	"fmt"
	"webbot/afforder/afforderstrategy"
	"webbot/afforder/afforderstrategy/functionafforder"
)

type AfforderStrategyID string

const (
	AfforderStrategyIDFunctionAfforder AfforderStrategyID = "function_afforder"
)

const DefaultAfforderStrategyID = AfforderStrategyIDFunctionAfforder

func AfforderStrategyByID(id AfforderStrategyID) (afforderstrategy.AfforderStrategy, error) {
	switch id {
	case AfforderStrategyIDFunctionAfforder:
		return functionafforder.New(), nil
	default:
		return nil, fmt.Errorf("invalid afforder strategy ID: %s", id)
	}
}
