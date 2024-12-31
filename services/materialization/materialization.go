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

			fmt.Printf("recieved request with image %s \n", imageName)

			query := `
				SELECT event_type 
				FROM events
				WHERE filename = $1
				ORDER BY timestamp ASC;
			`

			rows, err := db.Query(query, imageName)

			if err != nil {
				log.Fatal("Failed to make a query")
			}

			eventTypes := make([]string, 0)

			for rows.Next() {
				var eventType string

				err = rows.Scan(&eventType)
				if err != nil {
					log.Fatal("Faild to read row")
				}

				eventTypes = append(eventTypes, eventType)
				fmt.Printf("Read from DB event type %s \n", eventType)
			}
			rows.Close()

			startTime := time.Now()
			var computationSteps uint8 = 0

			for i := 0; i < len(eventTypes); i++ {
				fmt.Printf("current encoded steps %s \n", eventTypes[i])

				os.Rename(imageName, eventTypes[i]+imageName)
				imageName, err = OrchestratedCompute(imageName, ch, &computationSteps)
				if err != nil {
					log.Fatal("Failed to modify image")
				}

				fmt.Printf("recieved request with image %s \n", imageName)
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
