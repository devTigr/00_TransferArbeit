package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// failOnError is a helper function to log error messages
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type StockEvent struct {
	Company   string  `json:"company"`
	EventType string  `json:"eventType"`
	Price     float64 `json:"price"`
}

func main() {
	//Connect to DBCluster
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017/"))
	failOnError(err, "Failed to create a user for MongoDB")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	failOnError(err, "Failed to connect to MongoDB")

	defer client.Disconnect(ctx)

	//connect to RabbitMQ
	conn, err := amqp.Dial("amqp://stockmarket:supersecret123@localhost:5672/")

	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	queueName := os.Getenv("QUEUENAME")

	if queueName == "" {
		err := errors.New("queue missing")
		failOnError(err, "Failed to fetch the QUEUENAME from Environment")
	}

	q, err := ch.QueueDeclare(
		queueName, // name of the queue
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var events []StockEvent
	counter := 0

	for message := range msgs {
		var event StockEvent
		err := json.Unmarshal(message.Body, &event)
		if err != nil {
			log.Printf("Failed to unmarshal message: %s", err)
			continue
		}

		events = append(events, event)
		totalPrice := 0.0
		totalPrice += event.Price
		counter++

		if counter == 1000 {
			// Calculate the average price
			averagePrice := totalPrice / float64(counter)

			// Reset the events slice, counter, and total price
			events = []StockEvent{}
			counter = 0
			totalPrice = 0.0

			// Write the average price to MongoDB
			collection := client.Database("stockmarket").Collection("stocks")
			_, err := collection.InsertOne(ctx, bson.M{"name": queueName, "average_price": averagePrice})
			failOnError(err, "Failed to insert average price into MongoDB")
		}
	}

}
