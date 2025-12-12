package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Actype int

const (
	Ack Actype = iota
	NackRequeue
	NackDiscard
)

func Subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Actype,
	unmarshaller func([]byte) (T, error),
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return fmt.Errorf("error when subscribing json: %v", err)
	}
	newChan, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("error when consuming channel: %v", err)
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
				fmt.Printf(">")
			case NackDiscard:
				msg.Nack(false, false)
				fmt.Printf(">")
			case NackRequeue:
				msg.Nack(false, true)
				fmt.Printf(">")
			}
		}
	}()
	return nil
}
