# Provisioning Controller

A simplified Kubernetes controller for watching and managing provisioning repositories in Grafana.

## Overview

This controller watches `Repository` resources in Kubernetes and performs basic status management. It's a simplified version of the full provisioning controller that focuses on the core functionality of watching repositories and updating their status.

## Features

- Watches Repository resources in a specified namespace
- Updates repository status (observed generation)
- Graceful shutdown handling
- Configurable worker count
- Support for both in-cluster and external kubeconfig

## Building

```bash
# Build the binary
make build

# Clean build artifacts
make clean

# Install the binary
make install
```

## Running

### In-cluster (recommended for production)

```bash
# Run with default settings (watches 'default' namespace)
./bin/provisioning-controller

# Run with custom namespace
./bin/provisioning-controller --namespace=my-namespace

# Run with custom worker count
./bin/provisioning-controller --workers=4
```

### External cluster (for development)

```bash
# Run with kubeconfig file
./bin/provisioning-controller --kubeconfig=/path/to/kubeconfig

# Run with custom namespace and workers
./bin/provisioning-controller --kubeconfig=/path/to/kubeconfig --namespace=my-namespace --workers=2
```

## Command Line Options

- `--kubeconfig`: Path to kubeconfig file (optional, uses in-cluster config if not specified)
- `--namespace`: Namespace to watch (default: "default")
- `--workers`: Number of worker goroutines (default: 2)

## Architecture

The controller uses the standard Kubernetes controller pattern:

1. **Informer**: Watches Repository resources and maintains a local cache
2. **Work Queue**: Processes events in a rate-limited queue
3. **Workers**: Multiple goroutines process items from the queue
4. **Reconciliation**: Updates repository status when changes are detected

## Status Management

The controller currently manages:
- `status.observedGeneration`: Tracks the last processed generation of the repository spec

## Development

### Prerequisites

- Go 1.24+
- Kubernetes cluster with provisioning CRDs installed
- Access to the provisioning API

### Testing

```bash
# Run tests
make test
```

### Adding Features

This is a simplified version of the full provisioning controller. To add more functionality:

1. Extend the `processRepository` method to handle additional logic
2. Add new status fields to track additional state
3. Implement repository-specific operations (health checks, sync jobs, etc.)

## Integration with Full Controller

This controller can be used as a starting point for implementing the full provisioning controller functionality. The main differences are:

- Simplified status management (only observed generation)
- No repository-specific operations (health checks, sync, etc.)
- No complex dependency injection
- Focused on core watching and status updates

## Troubleshooting

### Common Issues

1. **Permission denied**: Ensure the controller has RBAC permissions to watch and update Repository resources
2. **CRD not found**: Make sure the provisioning CRDs are installed in the cluster
3. **Connection refused**: Verify kubeconfig or in-cluster configuration is correct

### Logs

The controller uses structured logging. Key log messages:

- `Starting SimpleRepositoryController`: Controller startup
- `Processing repository`: Processing a repository event
- `Updated repository status`: Status update completed
- `Received shutdown signal`: Graceful shutdown initiated

## License

This controller is part of the Grafana project and follows the same licensing terms. 