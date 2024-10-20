package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client
type Client struct {
	ID       int
	Messages int64
	Errors   int64
	Conn     *websocket.Conn
}

var errorCount int32
var totalMessages int64
var totalConnected int64

func main() {
	serverAddr := flag.String("server", "localhost:8081", "WebSocket server address")
	path := flag.String("path", "/average-price", "WebSocket server path")
	numClients := flag.Int("clients", 10000, "Number of concurrent clients")
	duration := flag.Int("duration", 10, "Duration of the test in seconds")

	flag.Parse()

	u := url.URL{Scheme: "ws", Host: *serverAddr, Path: *path}
	log.Printf("Connecting to %s", u.String())

	var connectWG sync.WaitGroup
	connectWG.Add(*numClients)

	var readWG sync.WaitGroup
	readWG.Add(*numClients)

	clients := make([]*Client, *numClients)

	startTime := time.Now()

	var clientsMu sync.Mutex

	connectionLimiter := make(chan struct{}, *numClients)

	go func() {
		ticker := time.NewTicker(1 * time.Millisecond)
		defer ticker.Stop()
		for i := 0; i < *numClients; i++ {
			<-ticker.C
			connectionLimiter <- struct{}{}
		}
	}()

	for i := 0; i < *numClients; i++ {
		go func(id int) {
			defer connectWG.Done()
			dialer := websocket.Dialer{
				HandshakeTimeout: 5 * time.Second,
			}

			<-connectionLimiter

			retryCount := 0
			maxRetries := 3
			var clientConn *websocket.Conn

			for {
				conn, _, err := dialer.Dial(u.String(), nil)
				if err != nil {
					if websocket.IsCloseError(err, websocket.CloseServiceRestart, websocket.CloseTryAgainLater) {
						log.Printf("Client %d: Server is busy. Retrying...", id)
						if retryCount < maxRetries {
							retryCount++
							time.Sleep(time.Duration(retryCount) * time.Second)
							continue
						} else {
							log.Printf("Client %d: Maximum retries reached. Giving up.", id)
							return
						}
					}

					atomic.AddInt32(&errorCount, 1)
					log.Printf("Client %d: Connection error: %v", id, err)
					return
				}

				retryCount = 0
				clientConn = conn

				break
			}

			client := &Client{
				ID:   id,
				Conn: clientConn,
			}

			clientsMu.Lock()
			clients[id] = client
			clientsMu.Unlock()

			atomic.AddInt64(&totalConnected, 1)

			go func(conn *websocket.Conn) {
				for {
					_, _, err := conn.ReadMessage()
					if err != nil {
						// log.Printf("Client %d: Read error: %v", id, err)
						atomic.AddInt64(&client.Errors, 1)
						return
					}
					atomic.AddInt64(&client.Messages, 1)
					atomic.AddInt64(&totalMessages, 1)
				}

			}(clientConn)
		}(i)
	}

	fmt.Println("Waiting for all clients to connect...")
	connectWG.Wait()

	connectionTime := time.Since(startTime)
	log.Printf("All clients connected in %v", connectionTime)

	log.Printf("Running test for %d seconds...", *duration)
	time.Sleep(time.Duration(*duration) * time.Second)

	totalTime := time.Since(startTime)

	log.Println("Closing all connections...")
	for _, client := range clients {
		if client != nil && client.Conn != nil {
			err := client.Conn.Close()
			if err != nil {
				log.Printf("Client %d: Error closing connection: %v", client.ID, err)
			}
		}
	}

	log.Println("Test completed")
	log.Printf("Total connection errors: %d", errorCount)
	log.Printf("Total clients attempted: %d", *numClients)
	log.Printf("Total clients connected: %d", totalConnected)
	log.Printf("Total messages received: %d", totalMessages)
	log.Printf("Total errors encountered: %d", errorCount)
	log.Printf("Total test duration: %v", totalTime)
	log.Printf("Average messages per client: %.2f", float64(totalMessages)/float64(totalConnected))
	log.Printf("Messages per second: %.2f", float64(totalMessages)/totalTime.Seconds())
}
