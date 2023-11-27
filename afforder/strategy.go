package afforder

import (
	"collaborativebrowser/afforder/afforderstrategy"
	"collaborativebrowser/afforder/afforderstrategy/functionafforder"
	"fmt"
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
