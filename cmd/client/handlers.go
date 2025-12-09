package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Actype {
	return func(ps routing.PlayingState) pubsub.Actype {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMessage(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.Actype {
	return func(move gamelogic.ArmyMove) pubsub.Actype {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(move)
		if moveOutcome == gamelogic.MoveOutComeSafe || moveOutcome == gamelogic.MoveOutcomeMakeWar {
			return pubsub.Ack
		} else {
			return pubsub.NackDiscard
		}
	}
}
