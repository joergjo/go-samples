# codeagent

This is a OpenAI port of Thorsten Ball's awesome code agent sample he has blogged about [here](https://ampcode.com/how-to-build-an-agent) using the [OpenAI Reponses API](https://platform.openai.com/docs/guides/text?api-mode=responses). I've opted to split the original sample across multiple Go files, but other than that this port reflects the original version based on Athropic's as closely as possible.

The application works with both Azure OpenAI and OpenAI and uses GPT-5-mini.

>When using Azure OpenAI, make sure to specify your base URL as documented [here](https://learn.microsoft.com/en-us/azure/ai-foundry/openai/api-version-lifecycle?tabs=go#v1-api-3). Note that this new endpoint does _not_ use the deployment name.  

Run the app as follows:
```
export OPENAI_BASE_URL=<your-base-url>
export OPENAI_API_KEY=<your-api-key>
go run .
```
