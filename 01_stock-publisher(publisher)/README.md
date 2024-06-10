# Stock Publisher

The `stock-publisher` is a Go application designed to simulate stock transaction events (buy or sell) for companies like Microsoft (MSFT), Tesla (TSLA), and Apple (AAPL). It connects to RabbitMQ and publishes synthetic stock price data to dedicated queues for each stock.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

What things you need to install the software and how to install them:

- Go (version 1.20 or later)
- Docker
- Access to a RabbitMQ server

### Installing

A step-by-step series of examples that tell you how to get a development environment running.

1. **Clone the repository**

   ```bash
   git clone https://github.com/switzerchees/stock-publisher.git
   cd stock-publisher
   ```

### Running

1. **Build the application**

   ```bash
   go build
   ```

2. **Run the application**

   ```bash
   ./stock-publisher
   ```

### Environment Variables

- `RABBITMQ_URL`: The URL of the RabbitMQ server (default: `amqp://stockmarket:supersecret123@localhost:5672/`)
- `TICKER_INTERVAL`: The interval in milliseconds between producing new stock price (default: `1`)
