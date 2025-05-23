package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/azure"
)

const apiVersion = "2024-10-21"

func main() {
	endpoint := os.Getenv("AZURE_OAI_ENDPOINT")
	key := os.Getenv("AZURE_OAI_KEY")
	deployment := os.Getenv("AZURE_OAI_DEPLOYMENT")

	if key == "" || deployment == "" || endpoint == "" {
		fmt.Println("Missing environment variables. Export AZURE_OAI_KEY, AZURE_OAI_DEPLOYMENT, and AZURE_OAI_ENDPOINT.")
		os.Exit(1)
	}

	client := openai.NewClient(azure.WithEndpoint(endpoint, apiVersion), azure.WithAPIKey(key))
	tools := []ToolDefinition{ReadFileDefinition, ListFilesDefinition, EditFileDefinition}

	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	agent := NewAgent(&client, deployment, getUserMessage, tools)
	err := agent.Run(context.TODO())
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}
