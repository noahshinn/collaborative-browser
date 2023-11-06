package binaryafforder

import (
	"encoding/json"
	"errors"
	"fmt"
	"webbot/utils/io"
)

func ParseDecisionTree(filepath string) (*DecisionNode, error) {
	var rootNode map[string]any
	if bytes, err := io.ReadFileAsBytes(filepath); err != nil {
		return nil, fmt.Errorf("failed to read bytes from file: %w", err)
	} else if err := json.Unmarshal(bytes, &rootNode); err != nil {
		return nil, fmt.Errorf("failed to unmarshall data: %w", err)
	} else {
		return parseDecisionTree(rootNode)
	}
}

func parseDecisionTree(node map[string]any) (*DecisionNode, error) {
	n, ok := node["name"]
	if !ok {
		return nil, errors.New("at least one node is missing a name")
	}
	name, ok := n.(string)
	if !ok {
		return nil, fmt.Errorf("name is type %T not string", n)
	}
	if value, ok := node["value"]; !ok {
		if s, ok := value.(string); !ok {
			return nil, fmt.Errorf("node with name %s has a func key with type %T", name, value)
		} else if s == "" {
			return nil, fmt.Errorf("node with name %s has an empty value for its func key", name)
		} else if _, ok := Libary[FuncKey(s)]; !ok {
			return nil, fmt.Errorf("node with name %s has a func key \"%s\" that was not found in the library", name, value)
		} else {
			return &DecisionNode{
				Typ:   DecisionTypeLeaf,
				Name:  name,
				Value: FuncKey(s),
			}, nil
		}
	} else {
		if _, ok := node["left"]; !ok {
			return nil, fmt.Errorf("non-root node with name %s does not have a left child", name)
		} else if leftData, ok := node["left"].(map[string]any); !ok {
			return nil, fmt.Errorf("non-root node with name %s does not have a valid left child data structure", name)
		} else if _, ok := node["right"]; !ok {
			return nil, fmt.Errorf("non-root node with name %s does not have a right child", name)
		} else if rightData, ok := node["right"].(map[string]any); !ok {
			return nil, fmt.Errorf("non-root node with name %s does not have a valid right child data structure", name)
		} else if i, ok := node["instruction"]; !ok {
			return nil, fmt.Errorf("non-root node with name %s is missing the instruction field", name)
		} else if instruction, ok := node["instruction"].(string); !ok {
			return nil, fmt.Errorf("non-root node with name %s has an instruction with type %T not string", name, i)
		} else {
			leftNode, err := parseDecisionTree(leftData)
			if err != nil {
				return nil, fmt.Errorf("failed to parse left node of node with name %s", name)
			}
			rightNode, err := parseDecisionTree(rightData)
			if err != nil {
				return nil, fmt.Errorf("failed to parse right node of node with name %s", name)
			}
			return &DecisionNode{
				Typ:         DecisionTypeNonLeaf,
				Name:        name,
				Left:        leftNode,
				Right:       rightNode,
				Instruction: instruction,
			}, nil
		}
	}
}
