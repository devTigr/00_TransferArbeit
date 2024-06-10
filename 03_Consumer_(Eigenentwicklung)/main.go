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
	// Umgebungsvariablen lesen
	mongoURL := os.Getenv("MONGODB_URL")
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	queueName := os.Getenv("QUEUENAME")

	if mongoURL == "" || rabbitMQURL == "" || queueName == "" {
		failOnError(errors.New("fehlende Umgebungsvariablen"), "Fehlende MONGODB_URL, RABBITMQ_URL oder QUEUENAME")
	}

	// Verbindung zu MongoDB Cluster
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) 
	defer cancel()
	
	clientOpts := options.Client().ApplyURI(mongoURL).SetServerSelectionTimeout(60 * time.Second)
	client, err := mongo.Connect(ctx, clientOpts)
	failOnError(err, "Failed to connect to MongoDB")
	defer client.Disconnect(ctx)
	

	// Verbindung zu RabbitMQ
	conn, err := amqp.Dial(rabbitMQURL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

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
	var totalPrice float64
	counter := 0

	for message := range msgs {
		var event StockEvent
		err := json.Unmarshal(message.Body, &event)
		if err != nil {
			log.Printf("Failed to unmarshal message: %s", err)
			continue
		}

		events = append(events, event)
		totalPrice += event.Price
		counter++

		if counter == 1000 {
			// Durchschnittspreis berechnen
			averagePrice := totalPrice / float64(counter)

			// Zurücksetzen des Events-Slice, Zähler und Gesamtpreis
			events = []StockEvent{}
			counter = 0
			totalPrice = 0.0

			// Durchschnittspreis in MongoDB speichern
			collection := client.Database("stockmarket").Collection("stocks")
			_, err := collection.InsertOne(ctx, bson.M{"company": queueName, "avgPrice": averagePrice})
			if err != nil {
				log.Printf("Failed to insert average price into MongoDB: %s", err)
			} else {
				log.Printf("Successfully inserted average price for %s into MongoDB", queueName)
			}
		}
	}
}
