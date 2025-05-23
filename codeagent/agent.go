package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
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
	fmt.Println("Chat with Azure OpenAI. Use CTRL-C to quit.")
	readUserInput := true
	var userInput, responseID string
	var input responses.ResponseNewParamsInputUnion

	for {
		if readUserInput {
			var ok bool
			fmt.Print("\u001b[94mYou\u001b[0m: ")
			userInput, ok = a.getUserMessage()
			if !ok {
				break
			}
			input.OfString = openai.String(userInput)
		}

		response, err := a.runInference(ctx, input, responseID)
		if err != nil {
			return err
		}
		responseID = response.ID
		input = responses.ResponseNewParamsInputUnion{}

		for _, output := range response.Output {
			switch output.Type {
			case "message":
				fmt.Printf("\u001b[93mOpenAI\u001b[0m: %s\n", response.OutputText())
			case "function_call":
				functionCall := output.AsFunctionCall()
				result, err := a.executeTool(functionCall.CallID, functionCall.Name, json.RawMessage(functionCall.Arguments))
				if err != nil {
					result = err.Error()
				}
				input.OfInputItemList = append(input.OfInputItemList, responses.ResponseInputItemParamOfFunctionCallOutput(functionCall.CallID, result))
			}
		}

		if len(input.OfInputItemList) == 0 {
			readUserInput = true
			continue
		}
		readUserInput = false
	}

	return nil
}

func (a *Agent) runInference(ctx context.Context, input responses.ResponseNewParamsInputUnion, responseID string) (*responses.Response, error) {
	tools := make([]responses.ToolUnionParam, 0, len(a.tools))
	for _, tool := range a.tools {
		tools = append(tools, responses.ToolUnionParam{
			OfFunction: &responses.FunctionToolParam{
				Name:        tool.Name,
				Description: param.NewOpt(tool.Description),
				Parameters:  tool.FunctionParams,
			},
		})
	}

	var previousResponseID param.Opt[string]
	if responseID != "" {
		previousResponseID = param.NewOpt(responseID)
	}
	message, err := a.client.Responses.New(ctx, responses.ResponseNewParams{
		Input:              input,
		MaxOutputTokens:    param.NewOpt(int64(1000)),
		Model:              shared.ChatModel(a.deployment),
		Tools:              tools,
		PreviousResponseID: previousResponseID,
	})
	return message, err
}

func (a *Agent) executeTool(id string, name string, input json.RawMessage) (string, error) {
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
		return "", fmt.Errorf("tool with ID %s not found", id)
	}

	fmt.Printf("\u001b[92mtool\u001b[0m: %s(%s)\n", name, input)
	response, err := toolDef.Function(input)
	if err != nil {
		return "", err
	}
	return response, nil
}
