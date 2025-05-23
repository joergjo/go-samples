package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/shared"
)

type Agent struct {
	client         *openai.Client
	deployment     string
	getUserMessage func() (string, bool)
	tools          []ToolDefinition
}

func NewAgent(client *openai.Client, deployment string, getUserMessage func() (string, bool), tools []ToolDefinition) *Agent {
	return &Agent{
		client:         client,
		deployment:     deployment,
		getUserMessage: getUserMessage,
		tools:          tools,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	conversation := []openai.ChatCompletionMessageParamUnion{}
	fmt.Println("Chat with Azure OpenAI. Use CTRL-C to quit.")
	readUserInput := true

	for {
		if readUserInput {
			fmt.Print("\u001b[94mYou\u001b[0m: ")
			userInput, ok := a.getUserMessage()
			if !ok {
				break
			}

			userMessage := openai.UserMessage(userInput)
			conversation = append(conversation, userMessage)
		}

		completion, err := a.runInference(ctx, conversation)
		if err != nil {
			return err
		}
		message := completion.Choices[0].Message
		conversation = append(conversation, message.ToParam())

		toolResults := []openai.ChatCompletionMessageParamUnion{}
		switch completion.Choices[0].FinishReason {
		case "stop":
			fmt.Printf("\u001b[93mOpenAI\u001b[0m: %s\n", message.Content)
		case "tool_calls":
			for _, tool := range message.ToolCalls {
				// fmt.Printf("Tool call: %s(%s)\n", tool.Function.Name, tool.Function.Arguments)
				result := a.executeTool(tool.ID, tool.Function.Name, json.RawMessage(tool.Function.Arguments))
				toolResults = append(toolResults, result)
			}
		}

		if len(toolResults) == 0 {
			readUserInput = true
			continue
		}
		readUserInput = false
		conversation = append(conversation, toolResults...)
	}

	return nil
}

func (a *Agent) runInference(ctx context.Context, conversation []openai.ChatCompletionMessageParamUnion) (*openai.ChatCompletion, error) {
	tools := make([]openai.ChatCompletionToolParam, 0, len(a.tools))
	for _, tool := range a.tools {
		tools = append(tools, openai.ChatCompletionToolParam{
			Function: shared.FunctionDefinitionParam{
				Name:        tool.Name,
				Description: param.NewOpt(tool.Description),
				Parameters:  tool.FunctionParams,
			},
		})
	}

	message, err := a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages:  conversation,
		MaxTokens: param.NewOpt(int64(1000)),
		Model:     shared.ChatModel(a.deployment),
		Tools:     tools,
	})
	return message, err
}

func (a *Agent) executeTool(id string, name string, input json.RawMessage) openai.ChatCompletionMessageParamUnion {
	var toolDef ToolDefinition
	var found bool
	for _, tool := range a.tools {
		if tool.Name == name {
			toolDef = tool
			found = true
			break
		}
	}
	if !found {
		return openai.ToolMessage("tool not found", id)
	}

	fmt.Printf("\u001b[92mtool\u001b[0m: %s(%s)\n", name, input)
	response, err := toolDef.Function(input)
	if err != nil {
		return openai.ToolMessage(err.Error(), id)
	}
	return openai.ToolMessage(response, id)
}
