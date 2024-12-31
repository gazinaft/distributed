// lab 3

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	dbname := os.Getenv("DB_NAME")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := "5432"

	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlconn)

	if err != nil {
		log.Fatalln("Couldn't connect to a DB")
	}

	defer db.Close()

	// check db
	err = db.Ping()
	if err != nil {
		log.Fatalln("DB does not respond")
	}
	fmt.Println("Connected!")

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
		"event_store", // name
		false,         // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
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
		log.Fatalf("Failed to register a consumer, %v", err)
	}
	var forever chan struct{}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		for d := range msgs {
			imageName := string(d.Body)

			fmt.Printf("recieved request with image %s \n", imageName)

			startTime := time.Now()

			encodedSteps := imageName[:8]

			insertDynStmt := `
				INSERT INTO events (filename, event_type, step, timestamp)
				VALUES ($1, $2, $3, $4);
			`
			// _, err := db.Query("select * from events;")

			// if err != nil {
			// 	log.Fatal("no events table")
			// }

			var computationSteps uint8 = 0

			for len(encodedSteps) > 0 && (encodedSteps[0] == 'C' || encodedSteps[0] == 'P') {
				// 		fmt.Printf("current encoded steps %s \n", encodedSteps)

				fmt.Printf("trying to insert into a DB: %s, %s, %d", imageName, encodedSteps[:1], computationSteps)

				_, err := db.Exec(insertDynStmt, imageName, encodedSteps[:1], computationSteps, time.Now())
				if err != nil {
					log.Fatalf("Failed to insert event into DB, %v", err)
				}

				encodedSteps = encodedSteps[1:]
			}

			err = ch.PublishWithContext(ctx,
				"",                // exchange
				"event_store_out", // routing key
				false,             // mandatory
				false,             // immediate
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          []byte(imageName),
					ReplyTo:       d.ReplyTo,
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
