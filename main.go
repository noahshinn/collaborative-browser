package main

import (
	"context"
	"fmt"
	"webbot/llm"
)

const testInputFile = "test.html"
const testOutputFile = "test.md"

func main() {
	// htmlText, err := utils.ReadFile(testInputFile)
	// if err != nil {
	// 	panic(err)
	// }
	// translator := &html2md.HTML2MDTranslater{}
	// translation, err := translator.Translate(htmlText)
	// if err != nil {
	// 	panic(fmt.Errorf("error translating: %w", err))
	// }
	// utils.WriteFile(testOutputFile, translation)

	ctx := context.Background()
	models := llm.AllModels
	model, err := models.ChatModel(string(llm.ChatModelGPT35Turbo))
	if err != nil {
		panic(err)
	}
	messages := []llm.Message{
		{
			Role:    llm.MessageRoleUser,
			Content: "hello",
		},
	}
	res, err := model.Message(ctx, messages, llm.MessageOptions{Temperature: 0.0})
	if err != nil {
		panic(err)
	}
	fmt.Println(res.Content)
}
