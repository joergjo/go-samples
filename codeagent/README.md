# codeagent

This is a OpenAI port of Thorsten Ball's awesome code agent sample he has blogged about [here](https://ampcode.com/how-to-build-an-agent). I've opted to split the original sample across multiple Go files, but other than that this port reflects the original version based on Athropic's as closely as possible [using the OpenAI Reponses API](https://platform.openai.com/docs/guides/text?api-mode=responses).

The code as written assumes the use of Azure OpenAI. To update the code to use OpenAI, change the options that are passed to the client constructor function:

Azure OpenAI
```go
client := openai.NewClient(azure.WithEndpoint(endpoint, apiVersion), azure.WithAPIKey(key))
```

OpenAI
```go
client := openai.NewClient(option.WithAPIKey("Your API Key")) // defaults to os.LookupEnv("OPENAI_API_KEY")
```

## Note
This sample uses [OpenAI's official Go SDK](https://github.com/openai/openai-go), _not_ [Microsoft's Azure OpenAI Go SDK](https://github.com/Azure/azure-sdk-for-go/tree/main/sdk/ai/azopenai).
