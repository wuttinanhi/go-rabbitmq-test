package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EXCHANGE_NAME = "go.test.exchange"
	QUEUE_NAME    = "hello_queue"
	MESSAGE       = "Hello World!"
)

func main() {
	mode := flag.String("mode", "default", "mode: producer or consumer")
	flag.Parse()

	conn, err := amqp.Dial("amqp://rabbitmquser:rabbitmqpassword@localhost:5672/")
	PanicIfError(err)
	defer conn.Close()

	channel, err2 := conn.Channel()
	PanicIfError(err2)
	defer channel.Close()

	err3 := channel.ExchangeDeclare(
		EXCHANGE_NAME, // exchange name
		"direct",      // exchange type (direct, fanout, topic, headers)
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	PanicIfError(err3)

	queue, err4 := channel.QueueDeclare(
		QUEUE_NAME, // queue name
		true,       // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	PanicIfError(err4)

	if *mode == "producer" {
		StartProducer(channel, &queue)
	} else if *mode == "consumer" {
		StartConsumer(channel, &queue)
	} else {
		fmt.Println("Please input mode: producer or consumer")
	}
}

func StartProducer(channel *amqp.Channel, queue *amqp.Queue) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	publish := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(MESSAGE),
	}
	err := channel.PublishWithContext(ctx,
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		publish,    // msg
	)
	PanicIfError(err)

	fmt.Printf("Sent -> %s\n", MESSAGE)
}

func StartConsumer(channel *amqp.Channel, queue *amqp.Queue) {
	msgs, err := channel.Consume(
		queue.Name,   // queue
		"consumer-1", // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	PanicIfError(err)

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
