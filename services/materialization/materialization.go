package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func OrchestratedCompute(filename string, ch *amqp.Channel, computationSteps *uint8) (res string, err error) {

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

	corrId := filename

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
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
			Priority:      *computationSteps, // higher prio for convolution
		})

	if err != nil {
		return "", err
	}

	*computationSteps += 1

	for d := range msgs {
		if corrId == d.CorrelationId {
			res = string(d.Body)
			break
		}
	}

	return
}

func main() {
	err_env := godotenv.Load(".env")

	if err_env != nil {
		log.Fatalf("Error loading .env file")
	}

	rabbit_user := os.Getenv("RABBIT_USER")
	rabbit_pw := os.Getenv("RABBIT_PASSWORD")

	rabbit_conn := fmt.Sprintf("amqp://%s:%s@rabbit:5672", rabbit_user, rabbit_pw)
	conn, err := amqp.Dial(rabbit_conn)

	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ")
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel")
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"event_store_out", // name
		false,             // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		log.Fatal("Failed to declare a queue")
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Fatal(err, "Failed to set QoS")
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatal("Failed to register a consumer")
	}
	var forever chan struct{}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		for d := range msgs {
			imageName := string(d.Body)
			encodedSteps := imageName[:8]

			fmt.Printf("recieved request with image %s \n", imageName)

			startTime := time.Now()
			var computationSteps uint8 = 0

			for len(encodedSteps) > 0 && (encodedSteps[0] == 'C' || encodedSteps[0] == 'P') {
				fmt.Printf("current encoded steps %s \n", encodedSteps)

				os.Rename(imageName, encodedSteps[:1]+imageName)
				imageName, err = OrchestratedCompute(imageName, ch, &computationSteps)
				if err != nil {
					log.Fatal("Failed to modify image")
				}

				fmt.Printf("recieved request with image %s \n", imageName)
				encodedSteps = encodedSteps[1:]
			}

			err = ch.PublishWithContext(ctx,
				"",        // exchange
				d.ReplyTo, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          []byte(imageName),
					Priority:      5,
				})

			if err != nil {
				log.Fatal("Failed to publish a message")
			}

			totalTime := time.Since(startTime)
			fmt.Printf("Took %d microseconds to handle request", totalTime.Microseconds())

			d.Ack(false)
		}
	}()

	fmt.Println("Awaiting RPC requests")
	<-forever
}
