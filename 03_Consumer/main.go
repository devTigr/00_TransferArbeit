package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type StockMsg struct {
	Company   string  `json:"company"`
	EventType string  `json:"eventType"`
	Price     float64 `json:"price"`
}

func checkError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("No value for %s in environment. Please check the documentation and values.", key)
	}
	return value
}

func connectMongo() *mongo.Collection {
	uri := getEnv("MONGODB_URL")
	clientOpts := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(context.TODO(), clientOpts)
	checkError(err, "MongoDB connection failed")

	err = client.Ping(context.TODO(), readpref.Primary())
	checkError(err, "MongoDB ping failed")

	return client.Database("stockmarket").Collection("stocks")
}

func avg(prices []float64) float64 {
	total := 0.0
	for _, price := range prices {
		total += price
	}
	return total / float64(len(prices))
}

func saveAvgToMongo(collection *mongo.Collection, avgPrice float64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc := bson.M{
		"company":  getEnv("QUEUE_NAME"),
		"avgPrice": avgPrice,
	}

	_, err := collection.InsertOne(ctx, doc)
	checkError(err, "Insert to MongoDB failed")
}

func processMessages(ch <-chan amqp.Delivery, collection *mongo.Collection, wg *sync.WaitGroup) {
	defer wg.Done()
	var prices []float64

	for msg := range ch {
		var stock StockMsg
		err := json.Unmarshal(msg.Body, &stock)
		if err != nil {
			log.Printf("Failed to unmarshal message: %s", err)
			continue
		}
		prices = append(prices, stock.Price)
	}

	if len(prices) > 0 {
		saveAvgToMongo(collection, avg(prices))
	}
}

func readStockMsgs(conn *amqp.Connection, collection *mongo.Collection) {
	ch, err := conn.Channel()
	checkError(err, "Channel opening failed")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		getEnv("QUEUE_NAME"),
		false, false, false, false, nil,
	)
	checkError(err, "Queue declaration failed")

	groupSize, err := strconv.Atoi(getEnv("MESSAGE_GROUP_SIZE"))
	checkError(err, "Group size parsing failed")

	msgs, err := ch.Consume(
		q.Name, "", true, false, false, false, nil,
	)
	checkError(err, "Message consumption failed")

	msgChannel := make(chan amqp.Delivery, groupSize)
	var wg sync.WaitGroup

	go func() {
		for msg := range msgs {
			msgChannel <- msg
			if len(msgChannel) == groupSize {
				close(msgChannel)
				wg.Add(1)
				go processMessages(msgChannel, collection, &wg)
				msgChannel = make(chan amqp.Delivery, groupSize)
			}
		}
		close(msgChannel)
		wg.Wait()
	}()
}

func main() {
	rabbitMQURL := getEnv("RABBITMQ_URL")
	conn, err := amqp.Dial(rabbitMQURL)
	checkError(err, "RabbitMQ connection failed")
	defer conn.Close()

	collection := connectMongo()

	go readStockMsgs(conn, collection)

	log.Println("Stock publishers are running. Press CTRL+C to exit.")
	select {}
}
