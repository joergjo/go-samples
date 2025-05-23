# A collection of Go based applications for fun and non-profit

![Gopher](https://github.com/joergjo/go-samples/assets/1625465/e2c80e13-4057-43ae-9027-0eb96b0f011f)

This repo contains various Go applications that I have ported from other development stacks (typically demo applications that I have found useful over the years) or small, yet useful applications I have developed myself. 
All samples use [Taskfiles](https://taskfile.dev) as simpler to use alternative to Makefiles and allow you to build the app, run tests, create a container image etc. Using Taskfiles isn't required to use these samples, but if 
you're new to Go will they help you to get started quickly without having to know the Go toolchain, the Docker CLI etc.

## Table of contents

1. [booklibrary](./booklibrary): This a port of Addy Osmani's venerable booklibrary API found in the Book Developing Backbone.js Applications originally written in JavaScript for Node.js.
   The Go version makes use of [chi](https://go-chi.io/#/) to implement as RESTful resource, the official [Mongo driver for Go](https://pkg.go.dev/go.mongodb.org/mongo-driver),
   and exposes [Prometheus metrics](https://prometheus.io/docs/tutorials/instrumenting_http_server_in_go/).

1. [kubeup](./kubeup): This is a webhook that listens for [Azure Kubernetes Service events received from Azure Event Grid](https://learn.microsoft.com/en-us/azure/aks/quickstart-event-grid?tabs=azure-cli) built using the
   [CloudEvents Go SDK](https://cloudevents.github.io/sdk-go/). This allows you to be notified of new Kubernetes versions becoming available, the start and end of rolling cluster upgrades or deprecation warnings if your
   Kubernetes will fall out of support. `kubeup` allows you to send forward these events by email, log them, and provide your own event handling logic. 

   Technically, you can achieve the same with an Azure Function or a Logic App,
   but being able to run a 10 MB container image on a [Azure Container App with scale-to-zero](https://learn.microsoft.com/en-us/azure/container-apps/scale-app?pivots=azure-cli#http) to me is a viable alternative.     

1. [ragserver](./ragserver/): This is a port of the RAG server sample that was published by the Go team on their official [blog](https://go.dev/blog/llmpowered) for Azure OpenAI/OpenAI.

1. [codeagent-cc](./codeagent-cc/): This is a port of Thorsten Ball's fantastic code editing agent sample published on his [blog](https://ampcode.com/how-to-build-an-agent) for Azure OpenAI/OpenAI using the [Chat Completions API](https://platform.openai.com/docs/guides/text?api-mode=chat).

1. [codeagent](./codeagent/): This is yet another port of Thorsten Ball's fantastic code editing agent sample published on his [blog](https://ampcode.com/how-to-build-an-agent) for Azure OpenAI/OpenAI using the [Responses API](https://platform.openai.com/docs/guides/text?api-mode=responses). The Respones API is more powerful than the Chat Completions API and allows one to maintain conversational state on the server side, which simplifies this implementation.