package actor

import (
	"context"
	"webbot/llm"
)

type StringActor interface {
	Act(ctx context.Context, messages []*llm.Message, functionDefs []*llm.FunctionDef) (string, error)
}
