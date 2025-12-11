package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Actype {
	return func(ps routing.PlayingState) pubsub.Actype {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMessage(gs *gamelogic.GameState, channel *amqp091.Channel) func(gamelogic.ArmyMove) pubsub.Actype {
	return func(move gamelogic.ArmyMove) pubsub.Actype {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(move)
		switch moveOutcome {
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				channel,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack

		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.Ack
		default:
			return pubsub.NackDiscard
		}
	}
}

func handlerWarMessages(gs *gamelogic.GameState, channel *amqp091.Channel) func(dw gamelogic.RecognitionOfWar) pubsub.Actype {
	return func(dw gamelogic.RecognitionOfWar) pubsub.Actype {
		defer fmt.Print("> ")
		warOutcome, winner, loser := gs.HandleWar(dw)
		switch warOutcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			resultMessage := fmt.Sprintf("%s won a war against %s", winner, loser)
			err := PublishGameLog(channel, gs.GetUsername(), resultMessage)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			resultMessage := fmt.Sprintf("%s won a war against %s", winner, loser)
			err := PublishGameLog(channel, gs.GetUsername(), resultMessage)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			resultMessage := fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			err := PublishGameLog(channel, gs.GetUsername(), resultMessage)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}
		fmt.Println("error: unknown war outcome")
		return pubsub.NackDiscard
	}
}

func PublishGameLog(ch *amqp091.Channel, username, message string) error {
	gamelog := routing.GameLog{
		Username:    username,
		CurrentTime: time.Now(),
		Message:     message,
	}
	publishingKey := routing.GameLogSlug + "." + username
	err := pubsub.PublishGob(ch, string(routing.ExchangePerilTopic), publishingKey, gamelog)
	if err != nil {
		return fmt.Errorf("error when publishing log: %d", err)
	}
	return nil
}
