package qron

import (
	"fmt"
	"strconv"
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
		return fmt.Errorf("failed to open a channel: %s", err)
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
			return fmt.Errorf("failed to declare an exchange: %s", err)
		}
	}

	q.channel = channel
	go func() {
		writeLog(lvlInfo, fmt.Sprintf("closing: %s ", <-connection.NotifyClose(make(chan *amqp.Error))))
		writeLog(lvlInfo, "trying to reconnect")
		for err := q.Connect(); err != nil; err = q.Connect() {
			writeLog(lvlError, err.Error())
			time.Sleep(5 * time.Second)
		}

	}()
	return nil
}

func (q *AMQP) Close() error {
	return q.channel.Close()
}

func (q *AMQP) Write(msgBody []byte, tags map[string]interface{}) error {
	q.Lock()
	defer q.Unlock()

	key := q.RoutingKey
	expiration := ""
	if ikey, ok := tags["key"]; ok {
		if skey, ok := ikey.(string); ok && skey != "" {
			key = skey
		}
	}
	if ittl, ok := tags["ttl"]; ok {
		if sttl, ok := ittl.(string); ok && sttl != "" {
			if ttl, err := time.ParseDuration(sttl); err == nil {
				expiration = strconv.FormatInt(ttl.Nanoseconds()/time.Millisecond.Nanoseconds(), 10)
			}
		} else if fttl, ok := ittl.(float64); ok && fttl > 0 {
			expiration = strconv.Itoa(int(fttl * 1000))
		}
	}

	err := q.channel.Publish(
		q.Exchange, // exchange
		key,        // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Expiration:  expiration,
			Timestamp:   time.Now(),
			Body:        msgBody,
		})
	if err != nil {
		return fmt.Errorf("failed to send amqp message: %s", err)
	}
	return nil
}
