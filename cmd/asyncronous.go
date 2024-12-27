package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	amqp "github.com/rabbitmq/amqp091-go"
)

func SendImageToServiceAsync(filename string) (res string, err error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

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

	corrId := filename

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
		})

	if err != nil {
		return "", err
	}

	for d := range msgs {
		if corrId == d.CorrelationId {
			res = string(d.Body)
			break
		}
	}

	return
}

func HandlePostImageAsync(c echo.Context) error {
	file, err := c.FormFile("image")

	if err != nil {
		return err
	}
	fmt.Println("Successfully submitted form")

	src, err := file.Open()
	if err != nil {
		return err
	}
	fmt.Println("Successfully opened file")

	defer src.Close()

	uuid, err := uuid.NewRandom()

	if err != nil {
		return err
	}

	// get unique filename
	filename := uuid.String() + filepath.Ext(file.Filename)
	fmt.Printf("created uuid %s \n", filename)

	// Destination
	fullFilePath := fmt.Sprintf("images/%s", filename)
	dst, err := os.Create(fullFilePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	alteredPic, err := SendImageToServiceAsync(filename)

	if err != nil {
		return err
	}

	htmlToImageFilePath := fmt.Sprintf("<img src=\"/images/%s\" id=\"returned-image\">", alteredPic)
	fmt.Println(htmlToImageFilePath)

	return c.HTML(http.StatusOK, htmlToImageFilePath)

}
