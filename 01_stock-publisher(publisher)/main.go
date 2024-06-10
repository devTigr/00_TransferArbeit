package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/streadway/amqp"
)

// failOnError is a helper function to log error messages
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// randomPrice generates a random price between $50 and $550
func randomPrice(r *rand.Rand) float64 {
	return r.Float64()*500 + 50
}

type StockEvent struct {
	Company   string  `json:"company"`
	EventType string  `json:"eventType"`
	Price     float64 `json:"price"`
}

// generate random buy/sell events for a stock
func stockPublisher(conn *amqp.Connection, stock string) {
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		stock, // name of the queue is the same as the stock
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	eventTypes := []string{"buy", "sell"}
	src := rand.NewSource(time.Now().UnixNano())
	localRand := rand.New(src) // Create a local random generator

	tickerIntervalValue := getEnvWithDefault("TICKER_INTERVAL", "1")
	tickerInterval, err := strconv.Atoi(tickerIntervalValue)
	failOnError(err, "Failed to parse ticker interval")

	ticker := time.NewTicker(time.Duration(tickerInterval) * time.Millisecond)
	for range ticker.C {
		eventType := eventTypes[rand.Intn(len(eventTypes))]
		price := randomPrice(localRand)
		stockEvent := StockEvent{
			Company:   stock,
			EventType: eventType,
			Price:     price,
		}

		jsonBody, err := json.Marshal(stockEvent)
		failOnError(err, "Error marshaling JSON")

		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        jsonBody,
			})
		failOnError(err, "Failed to publish a message")
		log.Printf(" [x] Sent %s", jsonBody)
	}
}

// getEnvWithDefault returns the value of an environment variable or a default value if the environment variable is not set
func getEnvWithDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func main() {
	rabbitMQConnectionString := getEnvWithDefault("RABBITMQ_URL", "amqp://stockmarket:supersecret123@localhost:5672/")

	conn, err := amqp.Dial(rabbitMQConnectionString)

	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	stocks := []string{"MSFT", "TSLA", "AAPL"}

	for _, stock := range stocks {
		go stockPublisher(conn, stock)
	}

	log.Printf("All stock publishers are running. Press CTRL+C to exit.")
	select {}
}
