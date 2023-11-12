package main

import (
	"bufio"
	"fmt"
	ampq "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
)

func main() {
	fmt.Println("Notification server start to work")
	defer fmt.Println("Notification server stop working")
	fmt.Println("Connecting to RMQ")
	conn, err := ampq.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		fmt.Println("Error in connecting RQM")
		fmt.Println(err)
		return
	}
	fmt.Println("Connecting to RMQ was successful")
	defer conn.Close()
	fmt.Println("Creating RMQ channel")
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println("Error in creating RMQ channel")
		fmt.Println(err)
		return
	}
	fmt.Println("Creating RMQ channel was successful")
	defer ch.Close()
	q, err := ch.QueueDeclare(
		"Inventory", // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		if text == "exit\n" {

			return
		}
	}
}
