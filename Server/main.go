package main

import (
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

func handler(w http.ResponseWriter, r *http.Request) {
    conn, _ := upgrader.Upgrade(w, r, nil)
    defer conn.Close()

    fmt.Printf("MT5 connected → %s\n", time.Now().Format("2006-01-02 15:04:05"))
    fmt.Println("Waiting for tick data...")

    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            fmt.Println("MT5 disconnected")
            break
        }

        // Beautiful live log
        fmt.Printf("%s → %s\n",
            time.Now().Format("15:04:05.000"),
            string(message))
    }
}

func main() {
    http.HandleFunc("/", handler)
    fmt.Println("Go Tick Receiver Running!")
    fmt.Println("Connect MT5 EA to: ws://127.0.0.1:7681")
    fmt.Println("Press Ctrl+C to stop")
    log.Fatal(http.ListenAndServe("127.0.0.1:7681", nil))
}