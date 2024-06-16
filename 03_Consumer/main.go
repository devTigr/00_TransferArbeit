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
	// URL wird aus den Umgebungsvariablen gelesen
	uri := getEnv("MONGODB_URL")
	// Mit options.Client aus dem MongoDB Driver packet wird die URL als option für clients konfiguriert
	clientOpts := options.Client().ApplyURI(uri)

	// der client wird erstellt und verbindung wird hergestellt. Hier wird nun die URL dem Client mittels der konfigurierten option mitgegeben.
	// context.TODO() ist dabei eine leeres Template um nicht alles spezifizieren zu müssen. clientOpts ist von uns mit URL befüllt.
	client, err := mongo.Connect(context.TODO(), clientOpts)
	checkError(err, "MongoDB connection failed")

	// DB-Cluster wird angepingt mit client.Ping() um sicherzustellen, dass eine Verbindung existiert
	// readpref.Primary() deklariert, dass der primary node des clusters angesprochenwerden soll.
	err = client.Ping(context.TODO(), readpref.Primary())
	checkError(err, "MongoDB ping failed")

	// ein Pointer zu den Daten (dieser Collection) als Struct wird retourniert
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
	log.Printf("Successfully inserted average price %f into MongoDB", avgPrice)
}

func processMessages(msgs []amqp.Delivery, collection *mongo.Collection) {
	var prices []float64

	// entpacken der msgs von RabbitMQ
	for _, msg := range msgs {
		var stock StockMsg
		err := json.Unmarshal(msg.Body, &stock)
		if err != nil {
			log.Printf("Failed to unmarshal message: %s", err)
			continue
		}
		// aggregieren der Preise
		prices = append(prices, stock.Price)
		log.Printf("Successfully read message from publisher: %s", msg.Body)
	}

	if len(prices) > 0 {
		// Speichern des durchschnittes in Mongo DB
		// wird zuvor mit avg(aggregiertePreise) berechnet.
		saveAvgToMongo(collection, avg(prices))
	}
	// acknowledge der messages nach dem Abarbeiten -> sollte hier nicht TRUE sein???
	for _, msg := range msgs {
		msg.Ack(false)
	}
}

func readStockMsgs(conn *amqp.Connection, collection *mongo.Collection, groupSize int) {
	// Channel wird geöffnet zum RabitMQ messagebroker
	ch, err := conn.Channel()
	checkError(err, "Channel opening failed")
	defer ch.Close()

	// Die Queue wird dazu definiert
	q, err := ch.QueueDeclare(
		getEnv("QUEUE_NAME"),
		false, false, false, false, nil,
	)
	checkError(err, "Queue declaration failed")

	//msgs wird erstellt und mit den Daten der RabbitMQ Queue befüllt
	msgs, err := ch.Consume(
		q.Name, "", false, false, false, false, nil,
	)
	checkError(err, "Message consumption failed")

	//whaitgroup zur synchronisation der read write prozesse wird erstellt
	var wg sync.WaitGroup
	// Batch
	msgBatch := make([]amqp.Delivery, 0, groupSize)
	batchLock := sync.Mutex{}

	go func() {
		for msg := range msgs {
			// sync.Mutex sperrt die Adressen für andere go routines (reserviert für diesen Prozess)
			batchLock.Lock()
			// alle Nachrichten werden zu dem Batch hinzugefügt
			msgBatch = append(msgBatch, msg)
			// wenn die groupSize erreicht ist wird der Batch verarbeitet sonst werden noch weitere Nachrichten hinzugefügt
			if len(msgBatch) >= groupSize {
				// ein counter zu der Whaitgroup hinzugefügt (alle in der wg warten bis dieser abgeschlossen.)
				wg.Add(1)
				// go routine zum verarbeiten der Nachrichten wird gestartet erhält am ende den Slice der msgs: msgBatch und die collection zum abspoeichern in der DB/collection
				go func(batch []amqp.Delivery) {
					// am ende dieser goroutine wird die whaitgroup wieder freigegeben.
					defer wg.Done()
					processMessages(batch, collection)
				}(msgBatch)
				//msgBatch wird geleert
				msgBatch = make([]amqp.Delivery, 0, groupSize)
			}
			// sync.Mutex wird wider freigegeben
			batchLock.Unlock()
		}
	}()

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt)
	<-c
	batchLock.Lock()
	if len(msgBatch) > 0 {
		processMessages(msgBatch, collection)
	}
	batchLock.Unlock()
	wg.Wait()
}

func main() {
	rabbitMQURL := getEnv("RABBITMQ_URL")
	conn, err := amqp.Dial(rabbitMQURL)
	checkError(err, "RabbitMQ connection failed")
	defer conn.Close()

	collection := connectMongo()

	groupSize, err := strconv.Atoi(getEnv("MESSAGE_GROUP_SIZE"))
	checkError(err, "Group size parsing failed")

	// go-Routine für das abholen der Daten vom RabbitMQ messageBroker
	go readStockMsgs(conn, collection, groupSize)

	log.Println("Stock publishers are running. Press CTRL+C to exit.")
	select {}
}
