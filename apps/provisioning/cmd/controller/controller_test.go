package main

import (
	"testing"

	"github.com/grafana/grafana-app-sdk/logging"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/cache"
)

func TestControllerCreation(t *testing.T) {
	// Test that we can create a controller instance
	controller := &SimpleRepositoryController{
		logger: logging.DefaultLogger.With("logger", "test"),
	}

	assert.NotNil(t, controller)
	assert.NotNil(t, controller.logger)
}

func TestKeyParsing(t *testing.T) {
	// Test that we can parse namespace/name keys correctly
	namespace, name, err := cache.SplitMetaNamespaceKey("default/test-repo")

	assert.NoError(t, err)
	assert.Equal(t, "default", namespace)
	assert.Equal(t, "test-repo", name)
}

func TestInvalidKeyParsing(t *testing.T) {
	// Test that invalid keys are handled correctly
	namespace, name, err := cache.SplitMetaNamespaceKey("invalid-key")

	// The function doesn't return an error, it just returns empty strings
	assert.NoError(t, err)
	assert.Equal(t, "", namespace)
	assert.Equal(t, "invalid-key", name)
}
