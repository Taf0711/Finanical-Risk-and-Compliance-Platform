# WebSocket Test Client

A standalone WebSocket test client for the Financial Risk Monitor platform.

## Usage

From the backend directory, run:

```bash
cd backend
go run ./tests/websocket-client/main.go
```

## What it does

- Connects to the WebSocket endpoint at `ws://localhost:8080/ws`
- Listens for real-time updates from the mock data generator
- Displays colored output for different message types:
  - 📈 Price updates (green)
  - ⚠️ Risk updates (yellow) 
  - 💸 Transactions (blue)
  - 🚨 Alerts (red)
- Shows live statistics every 10 messages and every 30 seconds
- Graceful shutdown with Ctrl+C

## Prerequisites

- Backend server must be running on port 8080
- Mock data generator should be enabled (development mode)
- Uses the backend's existing go.mod file (no separate dependencies needed)

## Example Output

```
🚀 Mock Data Generator Test Client
==================================================

Connecting to ws://localhost:8080/ws?user_id=test-client...
✅ Connected successfully!

[16:04:57] 👋 Welcome message received
[16:04:59] 📈 PRICE UPDATE #1
  AAPL: $150.25 ↑ +2.50%
  GOOGL: $2,800.00 ↓ -1.25%
```
