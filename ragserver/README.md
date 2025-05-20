# ragserver

This is a OpenAI port of the Go team's RAG server sample they have blogged about [here](https://go.dev/blog/llmpowered). I've added a [Docker Compose file](./compose.yaml) to run Weaviate and [HTTP test files](./tests/) to upload and query documents.  

The code as written assumes the use of Azure OpenAI. To update the code to use OpenAI, change the options that are passed to the client constructor function:

Azure OpenAI
```go
oaiClient, err := azopenai.NewClientWithKeyCredential(endpoint, creds, nil)
```

OpenAI
```go
oaiClient, err := azopenai.NewClientForOpenAI("https://api.openai.com/v1", creds, nil)
```

## Note
This sample uses [Microsoft's Azure OpenAI Go SDK](https://github.com/Azure/azure-sdk-for-go/tree/main/sdk/ai/azopenai), _not_ [OpenAI's official Go SDK](https://github.com/openai/openai-go).