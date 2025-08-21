package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

// ============================================
// MOCK DATA GENERATOR TEST CLIENT
// ============================================
// This is a standalone WebSocket test client
// Run with: go run main.go
// ============================================

type Stats struct {
	TotalMessages int
	PriceUpdates  int
	RiskUpdates   int
	Transactions  int
	Alerts        int
	StartTime     time.Time
}

type TestClient struct {
	conn  *websocket.Conn
	stats Stats
}

func main() {
	// Colors for output
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	fmt.Println(bold("🚀 Mock Data Generator Test Client"))
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	// Connect to WebSocket
	url := "ws://localhost:8080/ws?user_id=test-client"
	fmt.Printf("Connecting to %s...\n", url)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer conn.Close()

	fmt.Println(green("✅ Connected successfully!"))
	fmt.Println()

	client := &TestClient{
		conn: conn,
		stats: Stats{
			StartTime: time.Now(),
		},
	}

	// Handle Ctrl+C
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})

	// Read messages
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}

			var data map[string]interface{}
			if err := json.Unmarshal(message, &data); err != nil {
				log.Println("JSON error:", err)
				continue
			}

			client.stats.TotalMessages++
			msgType := data["type"].(string)
			timestamp := time.Now().Format("15:04:05")

			switch msgType {
			case "welcome":
				fmt.Printf("[%s] %s Welcome message received\n", timestamp, blue("👋"))

			case "price_update":
				client.stats.PriceUpdates++
				fmt.Printf("[%s] %s PRICE UPDATE #%d\n", timestamp, green("📈"), client.stats.PriceUpdates)
				if prices, ok := data["data"].(map[string]interface{}); ok {
					for symbol, priceData := range prices {
						if pd, ok := priceData.(map[string]interface{}); ok {
							price := pd["price"].(float64)
							change := pd["change"].(float64)
							arrow := "↑"
							changeColor := green
							if change < 0 {
								arrow = "↓"
								changeColor = red
							}
							fmt.Printf("  %s: $%.2f %s %s\n",
								symbol, price, arrow, changeColor(fmt.Sprintf("%.2f%%", change)))
						}
					}
				}

			case "risk_update":
				client.stats.RiskUpdates++
				fmt.Printf("[%s] %s RISK UPDATE #%d\n", timestamp, yellow("⚠️"), client.stats.RiskUpdates)
				if riskData, ok := data["data"].(map[string]interface{}); ok {
					if varData, ok := riskData["var"].(map[string]interface{}); ok {
						fmt.Printf("  VaR: $%.2f (Status: %s)\n",
							varData["Value"].(float64), varData["Status"])
					}
					if liqData, ok := riskData["liquidity"].(map[string]interface{}); ok {
						fmt.Printf("  Liquidity: %.2f%%\n",
							liqData["Value"].(float64)*100)
					}
				}

			case "new_transaction":
				client.stats.Transactions++
				fmt.Printf("[%s] %s TRANSACTION #%d\n", timestamp, blue("💸"), client.stats.Transactions)
				if txData, ok := data["data"].(map[string]interface{}); ok {
					if tx, ok := txData["transaction"].(map[string]interface{}); ok {
						fmt.Printf("  Type: %s | Symbol: %s | Amount: $%.2f\n",
							tx["TransactionType"], tx["Symbol"], tx["Amount"])
					}
				}

			case "new_alert", "aml_alert":
				client.stats.Alerts++
				fmt.Printf("[%s] %s ALERT #%d\n", timestamp, red("🚨"), client.stats.Alerts)
				if alertData, ok := data["data"].(map[string]interface{}); ok {
					if alert, ok := alertData["alert"].(map[string]interface{}); ok {
						fmt.Printf("  Title: %s\n", bold(alert["Title"].(string)))
						fmt.Printf("  Severity: %s\n", red(alert["Severity"].(string)))
						fmt.Printf("  Description: %s\n", alert["Description"])
					}
				}

			default:
				fmt.Printf("[%s] Unknown message type: %s\n", timestamp, msgType)
			}

			// Print statistics every 10 messages
			if client.stats.TotalMessages%10 == 0 {
				client.PrintStats()
			}
		}
	}()

	// Print stats every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			client.PrintStats()
		case <-interrupt:
			fmt.Println("\n" + yellow("Interrupt received, closing connection..."))

			// Final statistics
			fmt.Println("\n" + bold("📊 FINAL STATISTICS"))
			client.PrintStats()

			// Close connection
			err := conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Write close error:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func (c *TestClient) PrintStats() {
	elapsed := time.Since(c.stats.StartTime)
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Printf("📊 Statistics (Running for %s)\n", elapsed.Round(time.Second))
	fmt.Printf("  Total Messages: %d\n", c.stats.TotalMessages)
	fmt.Printf("  Price Updates:  %d (expect ~%.0f)\n",
		c.stats.PriceUpdates, elapsed.Seconds()/2)
	fmt.Printf("  Risk Updates:   %d (expect ~%.0f)\n",
		c.stats.RiskUpdates, elapsed.Seconds()/15)
	fmt.Printf("  Transactions:   %d (expect ~%.0f)\n",
		c.stats.Transactions, elapsed.Seconds()/10)
	fmt.Printf("  Alerts:         %d (expect ~%.0f)\n",
		c.stats.Alerts, elapsed.Seconds()/30)
	fmt.Println(strings.Repeat("-", 50) + "\n")
}
