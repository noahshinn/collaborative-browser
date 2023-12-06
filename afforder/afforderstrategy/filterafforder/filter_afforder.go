package filterafforder

import (
	"collaborativebrowser/afforder/afforderstrategy"
	"collaborativebrowser/afforder/afforderstrategy/functionafforder"
	"collaborativebrowser/browser"
	"collaborativebrowser/llm"
	"collaborativebrowser/trajectory"
	"collaborativebrowser/utils/slicesx"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type FilterAfforder struct {
	functionafforder.FunctionAfforder
	models *llm.Models
}

//go:embed system_prompt_to_filter_affordances.txt
var systemPromptToFilterAffordances string

func New(models *llm.Models) afforderstrategy.AfforderStrategy {
	return &FilterAfforder{
		models: models,
	}
}

func (fa *FilterAfforder) GetAffordances(ctx context.Context, traj *trajectory.Trajectory, br *browser.Browser) (messages []*llm.Message, functionDefs []*llm.FunctionDef, err error) {
	filteredPageRender, err := fa.filterBrowserDisplay(ctx, br, traj)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to render browser display: %w", err)
	}
	messages = fa.GetMessageAffordances(filteredPageRender, traj)
	functionDefs = fa.GetFunctionAffordances()
	return messages, functionDefs, nil
}

func (fa *FilterAfforder) filterBrowserDisplay(ctx context.Context, br *browser.Browser, traj *trajectory.Trajectory) (filteredBrowserDisplay string, err error) {
	rawBrowserDisplay := br.GetDisplay().MD
	numberedBrowserDisplay := displayBrowserContentWithLineno(rawBrowserDisplay)
	trajDisplay := traj.GetAbbreviatedText()
	messages := []*llm.Message{
		{
			Role:    llm.MessageRoleSystem,
			Content: systemPromptToFilterAffordances,
		},
		{
			Role: llm.MessageRoleUser,
			Content: fmt.Sprintf(`----- START BROWSER -----
%s
----- END BROWSER -----

----- START TRAJECTORY -----
%s
----- END TRAJECTORY -----

First, list a description of the next action that should be taken. Then, list a sequence of the irrelevant lines. These lines will be deleted and the remaining lines will be displayed as the web browser for your next action.`, numberedBrowserDisplay, trajDisplay),
		},
	}
	functionDef := &llm.FunctionDef{
		Name:        "filter_irrelevant_lines",
		Description: "Filter out the irrelevant lines from the browser display. The remaining lines will be displayed as the web browser for your next action.",
		Parameters: llm.Parameters{
			Type: "object",
			Properties: map[string]llm.Property{
				"next_action_description": {
					Type:        "string",
					Description: "A description of the next action that should be taken. This description should be a single line of text.",
				},
				"irrelevant_lines": {
					Type:        "array",
					Description: "A sequence of the irrelevant lines. Each item is a number or a range followed by a description.",
					Items: &llm.ArrayItems{
						Type: "string",
					},
				},
			},
			Required: []string{"next_action_description", "irrelevant_lines"},
		},
	}
	var args map[string]interface{}
	if res, err := fa.models.ChatModels[llm.ChatModelGPT4].Message(ctx, messages, &llm.MessageOptions{
		Temperature:  0.0,
		Functions:    []*llm.FunctionDef{functionDef},
		FunctionCall: "filter_affordances",
	}); err != nil {
		return "", fmt.Errorf("failed to get response from chat model: %w", err)
	} else if res.FunctionCall == nil {
		return "", fmt.Errorf("response from chat model did not include a function call")
	} else if res.FunctionCall.Name != "filter_affordances" {
		return "", fmt.Errorf("response from chat model did not include a function call to filter_affordances")
	} else if err := json.Unmarshal([]byte(res.FunctionCall.Arguments), &args); err != nil {
		return "", fmt.Errorf("failed to unmarshal arguments from chat model response: %w", err)
	} else if irrelevantLinesRes, ok := args["irrelevant_lines"].([]interface{}); !ok {
		return "", fmt.Errorf("failed to parse irrelevant lines from chat model response")
	} else if slicesx.Any(irrelevantLinesRes, func(item interface{}) bool {
		_, ok := item.(string)
		return !ok
	}) {
		return "", fmt.Errorf("failed to parse irrelevant lines from chat model response")
	} else if irrelevantLines, err := parseIrrelevantLinesResult(slicesx.Map(irrelevantLinesRes, func(item interface{}, i int) string {
		return item.(string)
	})); err != nil {
		return "", fmt.Errorf("failed to parse irrelevant lines from chat model response: %w", err)
	} else {
		relevantLines := slicesx.Filter(strings.Split(rawBrowserDisplay, "\n"), func(line string, i int) bool {
			_, ok := irrelevantLines[i]
			return !ok
		})
		return strings.Join(relevantLines, "\n"), nil
	}
}

func displayBrowserContentWithLineno(s string) string {
	lines := slicesx.Map(strings.Split(s, "\n"), func(line string, i int) string {
		return fmt.Sprintf("[%d] %s", i, line)
	})
	return strings.Join(lines, "\n")
}

func parseIrrelevantLinesResult(res []string) (map[int]struct{}, error) {
	irrelevantLines := make(map[int]struct{})
	re := regexp.MustCompile(`^(\d+(:|-)\d+|:\d+|\d+:|<\d+>) description="(.+)"$`)
	for _, line := range res {
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			return nil, fmt.Errorf("invalid format: %s", line)
		}
		var start, end int
		rangePart := matches[1]
		switch {
		case strings.Contains(rangePart, "-"):
			fmt.Sscanf(rangePart, "%d-%d", &start, &end)
		case strings.HasPrefix(rangePart, ":"):
			fmt.Sscanf(rangePart, ":%d", &end)
			start = 1 // Assuming start is 1 if not specified
		case strings.HasSuffix(rangePart, ":"):
			fmt.Sscanf(rangePart, "%d:", &start)
			end = -1 // End is unknown
		default:
			fmt.Sscanf(rangePart, "<%d>", &start)
			end = start
		}
		if start == end {
			irrelevantLines[start] = struct{}{}
		} else {
			for i := start; i < end; i++ {
				irrelevantLines[i] = struct{}{}
			}
		}
	}
	return irrelevantLines, nil
}
