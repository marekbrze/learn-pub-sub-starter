package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Actype int

const (
	Ack Actype = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Actype,
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("error when subscribing json: %v", err)
	}
	newChan, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("error when consuming channel: %v", err)
	}

	unmarshaller := func(data []byte) (T, error) {
		var target T
		err := json.Unmarshal(data, &target)
		return target, err
	}

	go func() {
		defer channel.Close()
		for msg := range newChan {
			target, err := unmarshaller(msg.Body)
			if err != nil {
				fmt.Printf("could not unmarshal message: %v\n", err)
				continue
			}
			actype := handler(target)
			switch actype {
			case Ack:
				msg.Ack(false)
				fmt.Println("Ack")
				fmt.Printf(">")
			case NackDiscard:
				msg.Nack(false, false)
				fmt.Println("Nack Discard")
				fmt.Printf(">")
			case NackRequeue:
				msg.Nack(false, true)
				fmt.Println("Nack Requeue")
				fmt.Printf(">")
			}
		}
	}()
	return nil
}
