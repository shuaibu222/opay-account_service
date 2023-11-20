package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SendToRabbitmq(queue string, data any) error {
	conn, err := rabbitmqConnection()
	if err != nil {
		log.Println("failed to create connection with rabbitmq", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		log.Println("failed to create channel", err)
	}

	defer conn.Close()
	defer channel.Close()

	q, err := channel.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		log.Println("failed to declare a queue", err)
	}

	// We use exchange when we want producer to send to different queues without interacting directly with queue
	err = channel.ExchangeDeclare(
		"account_exchange", // Exchange name
		"fanout",           // Exchange type
		true,               // Durable
		false,              // Auto-deleted
		false,              // Internal
		false,              // No-wait
		nil,                // Arguments
	)
	if err != nil {
		log.Println("Exchange declaration failed", err)
		return nil
	}

	// Bind the queue to the exchange to let them know each other. with that we can have as many queues as we want to the same exchange
	err = channel.QueueBind(
		q.Name,             // Queue name
		"",                 // Routing key
		"account_exchange", // Exchange
		false,
		nil,
	)
	if err != nil {
		log.Println("queue bind failed", err)
	}

	accountData, err := json.Marshal(data)
	if err != nil {
		log.Println("failed to marshal", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = channel.PublishWithContext(ctx,
		"account_exchange", // exchange
		"",                 // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(accountData),
		})

	if err != nil {
		log.Println("failed to publish with context", err)
	}

	log.Println("Successfully published")

	return nil
}

func rabbitmqConnection() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	// don't continue until rabbit is ready
	for {
		client, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			log.Println("RabbitMQ not yet ready...")
			counts++
		} else {
			log.Println("Connected to RabbitMQ!")
			connection = client
			break
		}

		if counts > 5 {
			log.Println(err)
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off...")
		time.Sleep(backOff)
		continue
	}

	return connection, nil
}
