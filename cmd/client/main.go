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
	queueNameKey := routing.PauseKey + "." + name
	_, _, err = pubsub.DeclareAndBind(connection, routing.ExchangePerilDirect, queueNameKey, routing.PauseKey, pubsub.SimpleQueueTransient)
	if err != nil {
		log.Fatal("Couldn't connect")
	}
	gameState := gamelogic.NewGameState(name)
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
			fmt.Println(moved)
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
