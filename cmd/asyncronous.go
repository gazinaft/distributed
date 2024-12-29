package main

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SendImageToServiceAsync(filename string) (res string, err error) {
	conn, err := amqp.Dial("amqp://guest:guest@rabbit:5672")

	if err != nil {
		return "", err
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return "", err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return "", err
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return "", err
	}

	fmt.Println("Connected to rabbit")

	corrId := filename

	var prio uint8

	if filename[0] == 'P' {
		prio = 1
	} else {
		prio = 5
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"",          // exchange
		"rpc_queue", // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrId,
			ReplyTo:       q.Name,
			Body:          []byte(filename),
			Priority:      prio, // higher prio for convolution
		})

	if err != nil {
		return "", err
	}

	fmt.Printf("Posted to to rabbit %s \n", filename)

	for d := range msgs {
		if corrId == d.CorrelationId {
			res = string(d.Body)
			break
		}
	}

	return
}

func SendImageToOrchestrator(filename string) (res string, err error) {
	conn, err := amqp.Dial("amqp://guest:guest@rabbit:5672")

	if err != nil {
		return "", err
	}

	fmt.Println("Connected to rabbit")

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return "", err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return "", err
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return "", err
	}

	fmt.Println("created correlation Id")

	corrId := filename

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"",                   // exchange
		"orchestrator_queue", // routing key
		false,                // mandatory
		false,                // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrId,
			ReplyTo:       q.Name,
			Body:          []byte(filename),
		})

	if err != nil {
		return "", err
	}
	fmt.Println("published request")

	for d := range msgs {
		if corrId == d.CorrelationId {
			res = string(d.Body)
			break
		}
	}

	return
}
