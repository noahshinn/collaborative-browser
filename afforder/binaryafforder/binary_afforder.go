package binaryafforder

import (
	"context"
	"webbot/afforder"
	"webbot/browser"
	"webbot/llm"
	"webbot/trajectory"
)

type BinaryAfforder struct {
}

func NewBinaryAfforder(chatModel llm.ChatModel) afforder.Afforder {
	return &BinaryAfforder{}
}

func (ba *BinaryAfforder) NextAction(ctx context.Context, traj *trajectory.Trajectory, br *browser.Browser) (nextAction trajectory.TrajectoryItem, render trajectory.TrajectoryItem, err error) {
	// TODO: implement
	return
}

type DecisionNode struct {
	Typ DecisionType

	// for non-leaf nodes
	Instruction string
	Left        *DecisionNode
	Right       *DecisionNode

	// for leaf-nodes
	Value FuncKey
}

type DecisionType string

const (
	DecisionTypeLeaf    DecisionType = "leaf"
	DecisionTypeNonLeaf DecisionType = "non-leaf"
)
