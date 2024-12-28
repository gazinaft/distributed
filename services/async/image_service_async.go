package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/gazinaft/distributed/util"
	"github.com/google/uuid"

	amqp "github.com/rabbitmq/amqp091-go"
)

func modifyImageAsync(filename string) (string, error) {

	fmt.Printf("ImagePath of original image %s \n", filename)

	img, err := util.GetImageFromFilePath(fmt.Sprintf("images/%s", filename))
	if err != nil {
		return "", err
	}

	uuid, err := uuid.NewRandom()

	if err != nil {
		return "", err
	}

	// get unique filename
	newFilename := uuid.String() + filepath.Ext(filename)
	fmt.Printf("created uuid %s \n", newFilename)

	newFilePath := fmt.Sprintf("images/%s", newFilename)

	err = util.WriteImageToFilePath(util.PosterizeImage(img, 5), newFilePath)

	if err != nil {
		return "", err
	}
	return newFilename, nil
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")

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
		"rpc_queue", // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
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

			fmt.Printf("recieved request with image %s \n", imageName)

			startTime := time.Now()

			response, err := modifyImageAsync(imageName)
			if err != nil {
				log.Fatal("Failed to modify image")
			}

			err = ch.PublishWithContext(ctx,
				"",        // exchange
				d.ReplyTo, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          []byte(response),
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
