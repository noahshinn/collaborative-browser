package binaryafforder

import (
	"context"
	"fmt"
	"webbot/afforder"
	"webbot/browser"
	"webbot/llm"
	"webbot/trajectory"
)

type BinaryAfforder struct {
	decisionRootNode    *DecisionNode
	decisionCallbackLib map[FuncKey]func() any
}

// TODO: add path
const decisionTreeFilepath = ""

func NewBinaryAfforder(chatModel llm.ChatModel) afforder.Afforder {
	decisionRootNode, err := ParseDecisionTree(decisionTreeFilepath)
	// TODO: figure this out
	if err != nil {
		panic(fmt.Errorf("error parsing decision tree: %w", err))
	}
	var decisionCallbackLib map[FuncKey]func() any
	return &BinaryAfforder{
		decisionRootNode:    decisionRootNode,
		decisionCallbackLib: decisionCallbackLib,
	}
}

func (ba *BinaryAfforder) NextAction(ctx context.Context, traj *trajectory.Trajectory, br *browser.Browser) (nextAction trajectory.TrajectoryItem, render trajectory.TrajectoryItem, err error) {
	// TODO: implement
	return
}

type DecisionNode struct {
	Typ  DecisionType
	Name string

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
