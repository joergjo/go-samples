# ragserver

This is a OpenAI port of the Go team's RAG server sample they have blogged about [here](https://go.dev/blog/llmpowered). I've added a [Docker Compose file](./compose.yaml) to run Weaviate and [HTTP test files](./tests/) to upload and query documents.  

The application works with both Azure OpenAI and OpenAI and uses GPT-4o-mini.

>When using Azure OpenAI, make sure to specify your base URL as documented [here](https://learn.microsoft.com/en-us/azure/ai-foundry/openai/api-version-lifecycle?tabs=go#v1-api-3). Note that this new endpoint does _not_ use the deployment name.  

Run the app as follows:
```
export OPENAI_BASE_URL=<your-base-url>
export OPENAI_API_KEY=<your-api-key>
go run .
```
