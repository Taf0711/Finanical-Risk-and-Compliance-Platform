#!/bin/bash

echo "🚀 Financial Risk Monitor - Backend Test Runner"
echo "=============================================="

# Check if server is running
echo "📋 Checking if server is running..."
if ! curl -s http://localhost:8080/api/v1/portfolios > /dev/null 2>&1; then
    echo "❌ Server is not running on localhost:8080"
    echo "🔧 Please start the server first with:"
    echo "   cd backend && go run ./cmd/api"
    exit 1
fi

echo "✅ Server is running!"
echo ""

# Run the test suite
echo "🧪 Running comprehensive test suite..."
cd tests
go run test_runner.go

echo ""
echo "🏁 Test run completed!"
