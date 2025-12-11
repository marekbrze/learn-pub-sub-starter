package pubsub

import (
	"bytes"
	"encoding/gob"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Actype,
) error {
	unmarshaller := func(data []byte) (T, error) {
		byteReader := bytes.NewReader(data)
		decoder := gob.NewDecoder(byteReader)
		var target T
		err := decoder.Decode(&target)
		if err != nil {
			return target, err
		}
		return target, nil
	}

	return Subscribe(conn, exchange, queueName, key, simpleQueueType, handler, unmarshaller)
}
