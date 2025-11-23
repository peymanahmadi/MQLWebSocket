// package main

// import (
//     "fmt"
//     "log"
//     "net/http"
//     "time"

//     "github.com/gorilla/websocket"
// )

// var upgrader = websocket.Upgrader{
//     CheckOrigin: func(r *http.Request) bool { return true },
// }

// func handler(w http.ResponseWriter, r *http.Request) {
//     conn, _ := upgrader.Upgrade(w, r, nil)
//     defer conn.Close()

//     fmt.Printf("MT5 connected → %s\n", time.Now().Format("2006-01-02 15:04:05"))
//     fmt.Println("Waiting for tick data...")

//     for {
//         _, message, err := conn.ReadMessage()
//         if err != nil {
//             fmt.Println("MT5 disconnected")
//             break
//         }

//         // Beautiful live log
//         fmt.Printf("%s → %s\n",
//             time.Now().Format("15:04:05.000"),
//             string(message))
//     }
// }

// func main() {
//     http.HandleFunc("/", handler)
//     fmt.Println("Go Tick Receiver Running!")
//     fmt.Println("Connect MT5 EA to: ws://127.0.0.1:7681")
//     fmt.Println("Press Ctrl+C to stop")
//     log.Fatal(http.ListenAndServe("127.0.0.1:7681", nil))
// }

package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strings"
    "time"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

// Base message structure
type BaseMessage struct {
    Type string `json:"type"`
}

// Tick data message
type TickMessage struct {
    Type   string  `json:"type"`
    Symbol string  `json:"symbol"`
    Bid    float64 `json:"bid"`
    Ask    float64 `json:"ask"`
    Time   string  `json:"time"`
    Spread int     `json:"spread"`
    Volume int64   `json:"volume"`
}

// Trade event message
type TradeMessage struct {
    Type        string  `json:"type"`
    Action      string  `json:"action"` // "open" or "close"
    Ticket      int64   `json:"ticket"`
    Symbol      string  `json:"symbol"`
    Direction   string  `json:"direction,omitempty"` // "buy" or "sell"
    Volume      float64 `json:"volume"`
    EntryPrice  float64 `json:"entry_price,omitempty"`
    ClosePrice  float64 `json:"close_price,omitempty"`
    Profit      float64 `json:"profit,omitempty"`
    Swap        float64 `json:"swap,omitempty"`
    Commission  float64 `json:"commission,omitempty"`
    TotalProfit float64 `json:"total_profit,omitempty"`
    SL          float64 `json:"sl,omitempty"`
    TP          float64 `json:"tp,omitempty"`
    OpenTime    string  `json:"open_time,omitempty"`
    CloseTime   string  `json:"close_time,omitempty"`
}

func handler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("Upgrade error: %v", err)
        return
    }
    defer conn.Close()

    fmt.Printf("\n✅ MT5 connected → %s\n", time.Now().Format("2006-01-02 15:04:05"))
    fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    fmt.Println("Listening for tick data and trade events...")
    fmt.Println()

    tickCount := 0
    tradeCount := 0

    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            fmt.Println("\n❌ MT5 disconnected")
            fmt.Printf("Session stats: %d ticks, %d trades\n", tickCount, tradeCount)
            break
        }

        // Parse base message to check type
        var base BaseMessage
        if err := json.Unmarshal(message, &base); err != nil {
            log.Printf("Error parsing message: %v", err)
            continue
        }

        timestamp := time.Now().Format("15:04:05.000")

        // Handle different message types
        switch base.Type {
        case "tick":
            var tick TickMessage
            if err := json.Unmarshal(message, &tick); err != nil {
                log.Printf("Error parsing tick: %v", err)
                continue
            }
            tickCount++
            fmt.Printf("[%s] 📊 TICK | %s | Bid: %.5f | Ask: %.5f | Spread: %d\n",
                timestamp, tick.Symbol, tick.Bid, tick.Ask, tick.Spread)

        case "trade":
            var trade TradeMessage
            if err := json.Unmarshal(message, &trade); err != nil {
                log.Printf("Error parsing trade: %v", err)
                continue
            }
            tradeCount++

            if trade.Action == "open" {
                emoji := "🟢"
                if trade.Direction == "sell" {
                    emoji = "🔴"
                }
                fmt.Printf("[%s] %s TRADE OPEN | Ticket: #%d | %s %s | %.2f lots @ %.5f",
                    timestamp, emoji, trade.Ticket, strings.ToUpper(trade.Direction), 
                    trade.Symbol, trade.Volume, trade.EntryPrice)
                
                if trade.SL > 0 || trade.TP > 0 {
                    fmt.Printf(" | SL: %.5f | TP: %.5f", trade.SL, trade.TP)
                }
                fmt.Println()
                
            } else if trade.Action == "close" {
                emoji := "✅"
                profitColor := ""
                if trade.TotalProfit < 0 {
                    emoji = "❌"
                    profitColor = "LOSS"
                } else {
                    profitColor = "PROFIT"
                }
                
                fmt.Printf("[%s] %s TRADE CLOSE | Ticket: #%d | %s | %.2f lots @ %.5f | %s: $%.2f",
                    timestamp, emoji, trade.Ticket, trade.Symbol, 
                    trade.Volume, trade.ClosePrice, profitColor, trade.TotalProfit)
                
                if trade.Swap != 0 || trade.Commission != 0 {
                    fmt.Printf(" (P: $%.2f, S: $%.2f, C: $%.2f)",
                        trade.Profit, trade.Swap, trade.Commission)
                }
                fmt.Println()
            }

        default:
            // Fallback for unknown message types
            fmt.Printf("[%s] 📩 %s\n", timestamp, string(message))
        }
    }
}

func main() {
    http.HandleFunc("/", handler)
    
    fmt.Println("╔═══════════════════════════════════════════════════════════╗")
    fmt.Println("║   Go WebSocket Server - Tick & Trade Event Streaming     ║")
    fmt.Println("╚═══════════════════════════════════════════════════════════╝")
    fmt.Printf("Endpoint: ws://127.0.0.1:7681\n")
    fmt.Printf("Started: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
    fmt.Println("⏳ Waiting for MT5 connection...")
    fmt.Println("Press Ctrl+C to stop\n")
    
    log.Fatal(http.ListenAndServe("127.0.0.1:7681", nil))
}