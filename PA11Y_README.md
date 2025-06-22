# Grafana Pa11y Accessibility Testing

This setup provides a Docker Compose environment to run Grafana server with automated accessibility testing using pa11y.

## Prerequisites

1. **Docker and Docker Compose** installed on your system
2. **Grafana built** - You need to build Grafana first to have the binary available

## Building Grafana

Before running the pa11y checks, you need to build Grafana:

```bash
# Build both backend and frontend
make build

# Or build just the backend
make build-go

# Or build just the frontend
make build-js
```

This will create the `bin/grafana-server` binary that the Docker Compose setup uses.

## Running Pa11y Checks

### Option 1: Using the convenience script

```bash
./scripts/run-pa11y.sh
```

### Option 2: Using Docker Compose directly

```bash
# Start the services
docker-compose -f docker-compose-pa11y.yaml up --build

# Or run in detached mode
docker-compose -f docker-compose-pa11y.yaml up -d --build

# View logs
docker-compose -f docker-compose-pa11y.yaml logs -f

# Stop the services
docker-compose -f docker-compose-pa11y.yaml down
```

## What the setup does

1. **Grafana Service**:

   - Runs Grafana server using the built binary from your host
   - Mounts configuration from `./conf`
   - Exposes Grafana on port 3001 (to avoid conflicts with local development)
   - Includes health checks to ensure Grafana is ready before pa11y runs

2. **Pa11y Service**:
   - Uses the existing `.pa11yci.conf.js` configuration
   - Waits for Grafana to be healthy before starting tests
   - Runs accessibility checks against multiple Grafana pages
   - Saves results to `./pa11y-results/pa11y-results.json`

## Configuration

The pa11y configuration is in `.pa11yci.conf.js` and includes tests for:

- Login page
- Dashboard pages
- Settings pages
- User management
- Plugin management
- And more...

## Results

After the tests complete, you can find the results in:

- `./pa11y-results/pa11y-results.json` - JSON format results
- Docker logs for detailed output

## Troubleshooting

### Grafana binary not found

If you get an error about the Grafana binary not being found:

```bash
make build
```

### Port conflicts

If port 3001 is already in use, you can modify the port mapping in `docker-compose-pa11y.yaml`:

```yaml
ports:
  - '3002:3000' # Change 3001 to another port
```

### Pa11y configuration

The pa11y configuration uses environment variables:

- `HOST=grafana` (the service name)
- `PORT=3000` (internal port)

These are automatically set in the Docker Compose file.

## Customization

You can modify the pa11y configuration by editing `.pa11yci.conf.js` or create a custom configuration file and update the volume mount in the Docker Compose file.
