#!/bin/bash

# Monte Carlo Risk Assessment Build Script
# Usage: ./run_monte_carlo.sh [options]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
VERBOSE=false
CLEAN=false
BUILD_ONLY=false
PROFILE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -c|--clean)
            CLEAN=true
            shift
            ;;
        -b|--build-only)
            BUILD_ONLY=true
            shift
            ;;
        -p|--profile)
            PROFILE=true
            shift
            ;;
        -h|--help)
            echo "Monte Carlo Risk Assessment Tool"
            echo ""
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  -v, --verbose     Enable verbose output"
            echo "  -c, --clean       Clean build before compilation"
            echo "  -b, --build-only  Build without running"
            echo "  -p, --profile     Enable profiling"
            echo "  -h, --help        Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                # Build and run with default settings"
            echo "  $0 -v -c          # Clean build with verbose output"
            echo "  $0 -b             # Build only, don't run"
            echo "  $0 -p             # Run with CPU profiling"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

echo -e "${BLUE}🎲 Monte Carlo Risk Assessment Tool${NC}"
echo -e "${BLUE}===================================${NC}"

# Check if we're in the correct directory
if [ ! -f "go.mod" ]; then
    echo -e "${RED}❌ Error: go.mod not found. Please run this script from the backend directory.${NC}"
    exit 1
fi

# Check Go installation
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Error: Go is not installed or not in PATH.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Go version: $(go version)${NC}"

# Clean build if requested
if [ "$CLEAN" = true ]; then
    echo -e "${YELLOW}🧹 Cleaning previous builds...${NC}"
    rm -f monte-carlo-tool
    go clean -cache
    go clean -modcache
fi

# Check dependencies
echo -e "${YELLOW}📦 Checking dependencies...${NC}"
if [ "$VERBOSE" = true ]; then
    go mod tidy -v
    go mod download -x
else
    go mod tidy
    go mod download
fi

# Create logs directory if it doesn't exist
mkdir -p logs

# Build the Monte Carlo tool
echo -e "${YELLOW}🔨 Building Monte Carlo tool...${NC}"
if [ "$VERBOSE" = true ]; then
    go build -v -ldflags="-s -w" -o monte-carlo-tool ./cmd/monte_carlo
else
    go build -ldflags="-s -w" -o monte-carlo-tool ./cmd/monte_carlo
fi

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Build successful!${NC}"
    
    # Get binary size
    BINARY_SIZE=$(du -h monte-carlo-tool | cut -f1)
    echo -e "${GREEN}📁 Binary size: ${BINARY_SIZE}${NC}"
else
    echo -e "${RED}❌ Build failed!${NC}"
    exit 1
fi

# Exit if build-only mode
if [ "$BUILD_ONLY" = true ]; then
    echo -e "${GREEN}🏁 Build complete. Exiting (build-only mode).${NC}"
    exit 0
fi

# Check if PostgreSQL and Redis are running (for integration tests)
echo -e "${YELLOW}🔍 Checking services...${NC}"

# Check if docker-compose services are running
if command -v docker-compose &> /dev/null; then
    if docker-compose ps | grep -q "postgres.*Up" && docker-compose ps | grep -q "redis.*Up"; then
        echo -e "${GREEN}✅ Database services are running${NC}"
    else
        echo -e "${YELLOW}⚠️  Database services not detected. Some features may be limited.${NC}"
    fi
fi

# Run the Monte Carlo simulation
echo -e "${YELLOW}🚀 Running Monte Carlo simulation...${NC}"

# Set environment variables
export MONTE_CARLO_LOG_LEVEL=INFO
if [ "$VERBOSE" = true ]; then
    export MONTE_CARLO_DEBUG=true
fi

# Create timestamp for this run
TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
LOG_FILE="logs/monte_carlo_${TIMESTAMP}.log"

echo -e "${BLUE}📝 Logging to: ${LOG_FILE}${NC}"

if [ "$PROFILE" = true ]; then
    # Run with profiling
    echo -e "${YELLOW}🔬 Running with CPU profiling...${NC}"
    go tool pprof -http=:8080 -seconds=30 ./monte-carlo-tool &
    PPROF_PID=$!
    
    # Run the tool and capture output
    if [ "$VERBOSE" = true ]; then
        ./monte-carlo-tool 2>&1 | tee "$LOG_FILE"
    else
        ./monte-carlo-tool > "$LOG_FILE" 2>&1
        cat "$LOG_FILE"
    fi
    
    # Stop profiling
    kill $PPROF_PID 2>/dev/null || true
    echo -e "${GREEN}📊 Profiling data available at http://localhost:8080${NC}"
else
    # Normal run
    if [ "$VERBOSE" = true ]; then
        ./monte-carlo-tool 2>&1 | tee "$LOG_FILE"
    else
        ./monte-carlo-tool > "$LOG_FILE" 2>&1
        cat "$LOG_FILE"
    fi
fi

EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ Monte Carlo simulation completed successfully!${NC}"
    
    # Display log file info
    LOG_SIZE=$(du -h "$LOG_FILE" | cut -f1)
    echo -e "${GREEN}📁 Log file size: ${LOG_SIZE}${NC}"
    
    # Extract key metrics from log
    echo -e "${BLUE}📊 Summary:${NC}"
    
    # Look for VaR results in log
    if grep -q "VaR_95" "$LOG_FILE"; then
        echo -e "${GREEN}💰 VaR Results Found:${NC}"
        grep "VaR_95\|VaR_99" "$LOG_FILE" | head -5
    fi
    
    # Look for performance metrics
    if grep -q "sims/sec" "$LOG_FILE"; then
        echo -e "${GREEN}⚡ Performance:${NC}"
        grep "sims/sec\|Duration\|Memory" "$LOG_FILE" | head -3
    fi
    
    # Look for validation results
    if grep -q "✅" "$LOG_FILE"; then
        echo -e "${GREEN}🔍 Validation Status:${NC}"
        grep "✅.*VaR\|✅.*convergence\|✅.*backtesting" "$LOG_FILE" | head -3
    fi
    
else
    echo -e "${RED}❌ Monte Carlo simulation failed with exit code: ${EXIT_CODE}${NC}"
    echo -e "${RED}📋 Check the log file for details: ${LOG_FILE}${NC}"
    
    # Show last few lines of log for quick debugging
    if [ -f "$LOG_FILE" ]; then
        echo -e "${YELLOW}🔍 Last 10 lines of log:${NC}"
        tail -10 "$LOG_FILE"
    fi
fi

# Clean up on exit
cleanup() {
    if [ "$PROFILE" = true ] && [ ! -z "$PPROF_PID" ]; then
        kill $PPROF_PID 2>/dev/null || true
    fi
}
trap cleanup EXIT

echo -e "${BLUE}🏁 Script completed.${NC}"
exit $EXIT_CODE
