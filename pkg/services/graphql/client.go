package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	contextmodel "github.com/grafana/grafana/pkg/services/contexthandler/model"
)

// APIClient handles requests to Grafana's /apis endpoints
type APIClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		httpClient: &http.Client{},
		baseURL:    strings.TrimSuffix(baseURL, "/"),
	}
}

// DashboardResource represents a dashboard from the /apis endpoint
type DashboardResource struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name            string            `json:"name"`
		Namespace       string            `json:"namespace"`
		UID             string            `json:"uid"`
		CreationTime    string            `json:"creationTimestamp"`
		Labels          map[string]string `json:"labels,omitempty"`
		Annotations     map[string]string `json:"annotations,omitempty"`
		ResourceVersion string            `json:"resourceVersion"`
	} `json:"metadata"`
	Spec struct {
		Title       string      `json:"title"`
		Description string      `json:"description,omitempty"`
		Tags        []string    `json:"tags,omitempty"`
		Dashboard   interface{} `json:"dashboard"` // The actual dashboard JSON
	} `json:"spec"`
	Status struct {
		URL string `json:"url,omitempty"`
	} `json:"status,omitempty"`
}

// DashboardList represents a list response from the dashboard API
type DashboardList struct {
	APIVersion string              `json:"apiVersion"`
	Kind       string              `json:"kind"`
	Items      []DashboardResource `json:"items"`
	Metadata   struct {
		Continue        string `json:"continue,omitempty"`
		ResourceVersion string `json:"resourceVersion"`
	} `json:"metadata"`
}

// GetDashboards retrieves dashboards from the /apis endpoint
func (c *APIClient) GetDashboards(ctx context.Context, namespace string, reqCtx *contextmodel.ReqContext) (*DashboardList, error) {
	// Build the URL
	apiPath := fmt.Sprintf("/apis/dashboard.grafana.app/v1beta1/namespaces/%s/dashboards", namespace)
	fullURL := c.baseURL + apiPath

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication headers from the request context
	if reqCtx != nil {
		// Copy authentication headers from the original request
		if reqCtx.Req != nil {
			if auth := reqCtx.Req.Header.Get("Authorization"); auth != "" {
				req.Header.Set("Authorization", auth)
			}
			// Copy cookies for session-based auth
			for _, cookie := range reqCtx.Req.Cookies() {
				req.AddCookie(cookie)
			}
		}
	}

	// Set appropriate headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var dashboardList DashboardList
	if err := json.NewDecoder(resp.Body).Decode(&dashboardList); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &dashboardList, nil
}

// GetDashboard retrieves a specific dashboard by name
func (c *APIClient) GetDashboard(ctx context.Context, namespace, name string, reqCtx *contextmodel.ReqContext) (*DashboardResource, error) {
	// Build the URL
	apiPath := fmt.Sprintf("/apis/dashboard.grafana.app/v1beta1/namespaces/%s/dashboards/%s", namespace, name)
	fullURL := c.baseURL + apiPath

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication headers from the request context
	if reqCtx != nil && reqCtx.Req != nil {
		if auth := reqCtx.Req.Header.Get("Authorization"); auth != "" {
			req.Header.Set("Authorization", auth)
		}
		for _, cookie := range reqCtx.Req.Cookies() {
			req.AddCookie(cookie)
		}
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var dashboard DashboardResource
	if err := json.NewDecoder(resp.Body).Decode(&dashboard); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &dashboard, nil
}

// Generic method for other resource types
func (c *APIClient) GetResources(ctx context.Context, group, version, namespace, resource string, reqCtx *contextmodel.ReqContext) (interface{}, error) {
	// Build the URL for generic resources
	var apiPath string
	if namespace != "" {
		apiPath = fmt.Sprintf("/apis/%s/%s/namespaces/%s/%s", group, version, namespace, resource)
	} else {
		apiPath = fmt.Sprintf("/apis/%s/%s/%s", group, version, resource)
	}

	fullURL := c.baseURL + apiPath

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication
	if reqCtx != nil && reqCtx.Req != nil {
		if auth := reqCtx.Req.Header.Get("Authorization"); auth != "" {
			req.Header.Set("Authorization", auth)
		}
		for _, cookie := range reqCtx.Req.Cookies() {
			req.AddCookie(cookie)
		}
	}

	req.Header.Set("Accept", "application/json")

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse as generic JSON
	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
