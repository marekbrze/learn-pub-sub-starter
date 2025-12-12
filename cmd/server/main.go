package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	fmt.Println("Starting Peril server...")

	connectionString := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatal("Couldn't connect")
	}
	defer connection.Close()

	fmt.Println("Connected to the server!")

	channel, err := connection.Channel()
	if err != nil {
		log.Fatal("Error when creating channel")
	}

	logChannel, _, err := pubsub.DeclareAndBind(connection, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", pubsub.SimpleQueueDurable)
	if err != nil {
		log.Fatal("Couldn't declare and bind GameLog queue", err)
	}

	err = pubsub.SubscribeGob(
		connection,
		routing.ExchangePerilTopic,
		string(routing.GameLogSlug),
		routing.GameLogSlug+".*",
		pubsub.SimpleQueueDurable,
		handleGameLog(logChannel),
	)
	if err != nil {
		log.Fatalf("could not subscribe to log queue: %v", err)
	}

	gamelogic.PrintServerHelp()
	for {
		input := gamelogic.GetInput()

		if len(input) == 0 {
			continue
		}
		firstWord := input[0]
		switch firstWord {
		case "pause":
			err = pubsub.PublishJSON(channel, string(routing.ExchangePerilDirect), string(routing.PauseKey), routing.PlayingState{IsPaused: true})
			if err != nil {
				log.Fatal("Error when publishing to the channel")
			}
		case "resume":
			err = pubsub.PublishJSON(channel, string(routing.ExchangePerilDirect), string(routing.PauseKey), routing.PlayingState{IsPaused: false})
			if err != nil {
				log.Fatal("Error when publishing to the channel")
			}
		case "quit":
			fmt.Println("Exiting")
			return
		default:
			fmt.Println("I don't understand. Please enter command again")
		}
	}
}
