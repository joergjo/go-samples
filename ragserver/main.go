// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command ragserver is an HTTP server that implements RAG (Retrieval
// Augmented Generation) using the Gemini model and Weaviate. See the
// accompanying README file for additional details.
package main

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

const systemPrompt = `
I will ask you a question and will provide some additional context information.
Assume this context information is factual and correct, as part of internal
documentation.
If the question relates to the context, answer it using the context.
If the question does not relate to the context, answer it as normal.

For example, let's say the context has nothing in it about tropical flowers;
then if I ask you about tropical flowers, just answer what you know about them
without referring to the context.

For example, if the context does mention minerology and I ask you about that,
provide information from the context along with general knowledge.
`

const ragTemplateStr = `
Question:
%s

Context:
%s
`

// This is a standard Go HTTP server. Server state is in the ragServer struct.
// The `main` function connects to the required services (Weaviate and OpenAI),
// initializes the server state and registers HTTP handlers.
func main() {
	endpoint := os.Getenv("AZURE_OAI_ENDPOINT")
	key := os.Getenv("AZURE_OAI_KEY")
	ccDeployment := os.Getenv("AZURE_OAI_CC_DEPLOYMENT")
	embDeployment := os.Getenv("AZURE_OAI_EMB_DEPLOYMENT")

	if key == "" || ccDeployment == "" || embDeployment == "" || endpoint == "" {
		slog.Error("missing environment variable")
		os.Exit(1)
	}

	ctx := context.Background()
	wvClient, err := initWeaviate(ctx)
	if err != nil {
		slog.Error("failed to initialize Weaviate client", "err", err)
		os.Exit(1)
	}

	creds := azcore.NewKeyCredential(key)
	oaiClient, err := azopenai.NewClientWithKeyCredential(endpoint, creds, nil)

	if err != nil {
		slog.Error("creating client", "error", err)
		os.Exit(1)
	}

	server := &ragServer{
		ctx:           ctx,
		wvClient:      wvClient,
		oaiClient:     oaiClient,
		ccDeployment:  ccDeployment,
		embDeployment: embDeployment,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /add/", server.addDocumentsHandler)
	mux.HandleFunc("POST /query/", server.queryHandler)

	port := cmp.Or(os.Getenv("SERVERPORT"), "9020")
	address := "localhost:" + port
	log.Println("listening on", address)
	log.Fatal(http.ListenAndServe(address, mux))
}

type ragServer struct {
	ctx           context.Context
	wvClient      *weaviate.Client
	oaiClient     *azopenai.Client
	ccDeployment  string
	embDeployment string
}

func (rs *ragServer) addDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse HTTP request from JSON.
	type document struct {
		Text string
	}
	type addRequest struct {
		Documents []document
	}
	ar := &addRequest{}

	err := readRequestJSON(r, ar)
	if err != nil {
		slog.Error("reading request", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input := make([]string, 0, len(ar.Documents))
	for _, d := range ar.Documents {
		input = append(input, d.Text)
	}
	resp, err := rs.oaiClient.GetEmbeddings(
		rs.ctx,
		azopenai.EmbeddingsOptions{
			DeploymentName: &rs.embDeployment,
			Input:          input,
		},
		nil)
	if err != nil {
		slog.Error("getting embeddings", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// Convert our documents - along with their embedding vectors - into types
	// used by the Weaviate client library.
	objects := make([]*models.Object, len(ar.Documents))
	for i, doc := range ar.Documents {
		objects[i] = &models.Object{
			Class: "Document",
			Properties: map[string]any{
				"text": doc.Text,
			},
			Vector: resp.Embeddings.Data[i].Embedding,
		}
	}

	// Store documents with embeddings in the Weaviate DB.
	log.Printf("storing %v objects in weaviate", len(objects))
	_, err = rs.wvClient.Batch().ObjectsBatcher().WithObjects(objects...).Do(rs.ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (rs *ragServer) queryHandler(w http.ResponseWriter, r *http.Request) {
	// Parse HTTP request from JSON.
	type queryRequest struct {
		Content string
	}
	qr := &queryRequest{}
	err := readRequestJSON(r, qr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Embed the query contents.
	embResp, err := rs.oaiClient.GetEmbeddings(
		rs.ctx,
		azopenai.EmbeddingsOptions{
			DeploymentName: &rs.embDeployment,
			Input:          []string{qr.Content},
		},
		nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Search weaviate to find the most relevant (closest in vector space)
	// documents to the query.
	gql := rs.wvClient.GraphQL()
	result, err := gql.Get().
		WithNearVector(
			gql.NearVectorArgBuilder().WithVector(embResp.Data[0].Embedding)).
		WithClassName("Document").
		WithFields(graphql.Field{Name: "text"}).
		WithLimit(3).
		Do(rs.ctx)
	if err := combinedWeaviateError(result, err); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	contents, err := decodeGetResults(result)
	if err != nil {
		http.Error(w, fmt.Errorf("reading weaviate response: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	// Create a RAG query for the LLM with the most relevant documents as
	// context.
	ragQuery := fmt.Sprintf(ragTemplateStr, qr.Content, strings.Join(contents, "\n"))
	messages := []azopenai.ChatRequestMessageClassification{
		// You set the tone and rules of the conversation with a prompt as the system role.
		&azopenai.ChatRequestSystemMessage{Content: azopenai.NewChatRequestSystemMessageContent(systemPrompt)},

		// The user asks a question
		&azopenai.ChatRequestUserMessage{Content: azopenai.NewChatRequestUserMessageContent(ragQuery)},
	}

	ccResp, err := rs.oaiClient.GetChatCompletions(rs.ctx, azopenai.ChatCompletionsOptions{
		DeploymentName: &rs.ccDeployment,
		Messages:       messages,
	}, nil)
	if err != nil {
		log.Printf("calling generative model: %v", err.Error())
		http.Error(w, "generative model error", http.StatusInternalServerError)
		return
	}

	if len(ccResp.Choices) != 1 {
		log.Printf("got %v candidates, expected 1", len(ccResp.Choices))
		http.Error(w, "generative model error", http.StatusInternalServerError)
		return
	}

	renderJSON(w, *ccResp.Choices[0].Message.Content)
}

// decodeGetResults decodes the result returned by Weaviate's GraphQL Get
// query; these are returned as a nested map[string]any (just like JSON
// unmarshaled into a map[string]any). We have to extract all document contents
// as a list of strings.
func decodeGetResults(result *models.GraphQLResponse) ([]string, error) {
	data, ok := result.Data["Get"]
	if !ok {
		return nil, fmt.Errorf("get key not found in result")
	}
	doc, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("get key unexpected type")
	}
	slc, ok := doc["Document"].([]any)
	if !ok {
		return nil, fmt.Errorf("document is not a list of results")
	}

	var out []string
	for _, s := range slc {
		smap, ok := s.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid element in list of documents")
		}
		s, ok := smap["text"].(string)
		if !ok {
			return nil, fmt.Errorf("expected string in list of documents")
		}
		out = append(out, s)
	}
	return out, nil
}
