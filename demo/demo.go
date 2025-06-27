package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/grafana/grafana/pkg/services/graphql"
)

func main() {
	// Create GraphQL service
	service, err := graphql.NewService()
	if err != nil {
		panic(err)
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(service.HandleGraphQL))
	defer server.Close()

	// Test query
	query := `{
		resources(group: "apps", version: "v1", resource: "deployments") {
			name
			kind
			namespace
		}
	}`

	// Create request
	requestBody := map[string]interface{}{
		"query": query,
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Make request
	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Print response
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Println("GraphQL Response:")
	prettyJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(prettyJSON))
}
