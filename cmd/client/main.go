package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	connectionString := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatal("Couldn't connect")
	}
	defer connection.Close()

	fmt.Println("Connected to the server!")
	name, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal("Couldn't connect")
	}
	queueNameKey := routing.PauseKey + "." + name
	_, queue, err := pubsub.DeclareAndBind(connection, routing.ExchangePerilDirect, queueNameKey, routing.PauseKey, pubsub.SimpleQueueTransient)
	if err != nil {
		log.Fatal("Couldn't connect")
	}
	fmt.Println(queue.Name)
	// Closing sequence
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Println("Connection closing")
	connection.Close()
}
