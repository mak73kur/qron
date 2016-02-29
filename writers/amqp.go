package writers

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type AMQP struct {
	// AMQP brokers
	URL string
	// AMQP exchange
	Exchange string
	// Routing Key Tag
	RoutingKey string

	channel *amqp.Channel
	sync.Mutex
}

func NewAMQP(url, exchange, routingKey string) (*AMQP, error) {
	p := &AMQP{URL: url, Exchange: exchange, RoutingKey: routingKey}
	err := p.Connect()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (q *AMQP) Connect() error {
	q.Lock()
	defer q.Unlock()

	var connection *amqp.Connection

	connection, err := amqp.Dial(q.URL)
	if err != nil {
		return err
	}
	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("Failed to open a channel: %s", err)
	}

	if q.Exchange != "" {
		err = channel.ExchangeDeclare(
			q.Exchange, // name
			"direct",   // type
			true,       // durable
			false,      // delete when unused
			false,      // internal
			false,      // no-wait
			nil,        // arguments
		)
		if err != nil {
			return fmt.Errorf("Failed to declare an exchange: %s", err)
		}
	}

	q.channel = channel
	go func() {
		log.Printf("Closing: %s", <-connection.NotifyClose(make(chan *amqp.Error)))
		log.Printf("Trying to reconnect")
		for err := q.Connect(); err != nil; err = q.Connect() {
			log.Println(err)
			time.Sleep(5 * time.Second)
		}

	}()
	return nil
}

func (q *AMQP) Close() error {
	return q.channel.Close()
}

func (q *AMQP) Write(msgBody []byte) error {
	q.Lock()
	defer q.Unlock()

	err := q.channel.Publish(
		q.Exchange,   // exchange
		q.RoutingKey, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msgBody,
		})
	if err != nil {
		return fmt.Errorf("Failed to send amqp message: %s", err)
	}
	return nil
}
