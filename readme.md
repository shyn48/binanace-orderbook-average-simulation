# Binance Orderbook Average Simulation

This Go script simulates a WebSocket client that connects to a Binance order book, collects messages, and logs various statistics such as total messages received, total errors encountered, and average messages per client.

## Prerequisites

- Go 1.22.3 or later
- Internet connection

## Installation

1. Clone the repository:
   ```sh
   git clone https://github.com/yourusername/binance-orderbook-average-simulation.git
   ```
2. Navigate to the project directory:
   ```sh
   cd binance-orderbook-average-simulation
   ```
3. Install dependencies:
   ```sh
   go mod tidy
   ```

## Usage

To run the script, use the following command:

```sh
go run main.go
```

## Code Overview

### main.go

The main script initializes WebSocket clients, collects messages, and logs statistics.

#### Key Structures and Variables

- `Client`: Represents a WebSocket client with fields for ID, message count, error count, and connection.
- `errorCount`: Tracks the total number of connection errors.
- `totalMessages`: Tracks the total number of messages received.
- `totalConnected`: Tracks the total number of clients successfully connected.

#### Logging

The script logs various statistics at the end of the test, including:

- Total connection errors
- Total clients attempted
- Total clients connected
- Total messages received
- Total errors encountered
- Total test duration
- Average messages per client
- Messages per second

### Dependencies

The project uses the `gorilla/websocket` package for WebSocket communication. The dependency is specified in the `go.mod` and `go.sum` files.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.