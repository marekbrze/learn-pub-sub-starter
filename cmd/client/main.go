package main

import (
	"fmt"
	"log"

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

	publishCh, err := connection.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	gameState := gamelogic.NewGameState(name)
	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gameState.GetUsername(),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gameState),
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	queueNameForUser := "army_moves." + name
	_, _, err = pubsub.DeclareAndBind(connection, routing.ExchangePerilTopic, queueNameForUser, "army_moves.*", pubsub.SimpleQueueTransient)
	if err != nil {
		log.Fatalf("coulnd't create queue for publishing moves: %v", err)
	}

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gameState.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMessage(gameState, publishCh),
	)
	if err != nil {
		log.Fatalf("could not subscribe to army moves: %v", err)
	}
	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQueueDurable,
		handlerWarMessages(gameState),
	)
	if err != nil {
		log.Fatalf("could not subscribe to war declarations: %v", err)
	}
	for {
		input := gamelogic.GetInput()

		if len(input) == 0 {
			continue
		}
		fmt.Println(gameState.GetUsername())
		firstWord := input[0]
		switch firstWord {
		case "spawn":
			err := gameState.CommandSpawn(input)
			if err != nil {
				fmt.Printf("couldn't process spawn command: %v\n", err)
				continue
			}
		case "move":
			moved, err := gameState.CommandMove(input)
			if err != nil {
				fmt.Printf("couldn't process move command: %v\n", err)
				continue
			}

			err = pubsub.PublishJSON(publishCh, string(routing.ExchangePerilTopic), queueNameForUser, moved)
			if err != nil {
				log.Fatal("Error when publishing to the channel")
			}

		case "status":
			gameState.CommandStatus()
			continue
		case "help":
			gamelogic.PrintClientHelp()
			continue
		case "spam":
			fmt.Println("Spamming not allowed yet!")
			continue
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("I don't understand. Please enter command again")
		}
	}
}
