#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting Grafana with pa11y accessibility checks...${NC}"

# Check if Grafana binary exists
if [ ! -f "./bin/grafana-server" ]; then
    echo -e "${RED}Error: Grafana binary not found at ./bin/grafana-server${NC}"
    echo -e "${YELLOW}Please build Grafana first using: make build${NC}"
    exit 1
fi

# Create results directory
mkdir -p pa11y-results

# Stop any existing containers
echo -e "${YELLOW}Stopping any existing containers...${NC}"
docker-compose -f docker-compose-pa11y.yaml down

# Start the services
echo -e "${YELLOW}Starting Grafana and pa11y services...${NC}"
docker-compose -f docker-compose-pa11y.yaml up --build

echo -e "${GREEN}Pa11y checks completed!${NC}"
echo -e "${YELLOW}Results are available in: ./pa11y-results/pa11y-results.json${NC}"
