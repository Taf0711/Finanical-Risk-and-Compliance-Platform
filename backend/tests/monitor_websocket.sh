#!/bin/bash

# ============================================
# WEBSOCKET MOCK DATA MONITOR - TERMINAL
# ============================================

echo "🚀 Financial Risk Monitor - Mock Data Terminal Monitor"
echo "======================================================"
echo ""

# Check if wscat is installed
if ! command -v wscat &> /dev/null; then
    echo "Installing wscat..."
    npm install -g wscat
fi

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "Please install jq for JSON formatting:"
    echo "  macOS: brew install jq"
    echo "  Linux: sudo apt-get install jq"
    exit 1
fi

echo "📡 Connecting to WebSocket..."
echo ""

# Connect and format output
wscat -c ws://localhost:8080/ws?user_id=terminal-monitor | while read -r line; do
    # Parse JSON and format based on type
    msg_type=$(echo "$line" | jq -r '.type' 2>/dev/null)
    timestamp=$(date +"%H:%M:%S")
    
    case "$msg_type" in
        "welcome")
            echo "[$timestamp] 👋 CONNECTED"
            echo "$line" | jq '.data'
            echo "---"
            ;;
        "price_update")
            echo "[$timestamp] 📈 PRICE UPDATE"
            echo "$line" | jq '.data | to_entries[] | "\(.key): $\(.value.price | tostring) (\(.value.change | tostring)%)"'
            echo "---"
            ;;
        "risk_update")
            echo "[$timestamp] ⚠️ RISK METRICS"
            echo "$line" | jq '.data.var, .data.liquidity'
            echo "---"
            ;;
        "new_alert"|"aml_alert")
            echo "[$timestamp] 🚨 ALERT"
            echo "$line" | jq '.data.alert | "Title: \(.Title)\nSeverity: \(.Severity)\nDescription: \(.Description)"'
            echo "---"
            ;;
        "new_transaction")
            echo "[$timestamp] 💸 TRANSACTION"
            echo "$line" | jq '.data.transaction | "Type: \(.TransactionType)\nSymbol: \(.Symbol)\nAmount: $\(.Amount)"'
            echo "---"
            ;;
        *)
            echo "[$timestamp] 📊 $msg_type"
            echo "$line" | jq '.'
            echo "---"
            ;;
    esac
done