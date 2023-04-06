package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// Declaring constant
const QUEUE_NAME = "github"
const CONFIG_PATH = "/config/config.yaml"

// Error message
func FailOnError(err error, msg string) error {
	if err != nil {
		log.Errorf(msg, err)
		return nil
	}
	return nil
}

func consumeMessages(queueName string, msgs <-chan amqp.Delivery) error {

	log.Info("Started listening to the queue " + queueName + " ... ")
	wg := new(sync.WaitGroup)
	for msg := range msgs {
		log.Info("Received message")

		// Unmarshal message body as JSON
		var bodyData map[string]interface{}
		err := json.Unmarshal(msg.Body, &bodyData)
		if err != nil {
			log.Info("Error unmarshalling message body: ", err)
			continue // skip this message and move to the next one
		}

		// Seperating header and body
		header := bodyData["header"]
		body := bodyData["body"]
		wg.Add(1)
		go Parser(wg, header, body)
	}
	wg.Wait()
	return nil
}

func main() {
	user := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")

	// Reading YAML files
	yamlFile, err := ioutil.ReadFile(CONFIG_PATH)
	FailOnError(err, "Error reading YAML file")

	// Unmarshalling the yaml file into the yamlconfig
	var yamlConfig map[string]string
	err = yaml.Unmarshal(yamlFile, &yamlConfig)
	FailOnError(err, "Failed to parse YAML file")

	// Connecting to RabbitMQ
	conn, err := amqp.Dial("amqp://" + user + ":" + password + "@" + yamlConfig["rabbitmq_entrypoint"] + ":" + yamlConfig["rabbitmq_portno"])
	FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// Connecting to the channel in RabbitMQ
	ch, err := conn.Channel()
	FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	msgs := make(<-chan amqp.Delivery, 1)
	msgs, err = ch.Consume(
		QUEUE_NAME, // queue name
		"",         // consumer name
		true,       // auto-acknowledge messages
		false,      // exclusive consumer
		false,      // no local consumer
		false,      // no wait
		nil,        // arguments
	)
	if err != nil {
		log.Fatalf("Error in creating channel for queue %s: %v", QUEUE_NAME, err)
		return
	}
	FailOnError(err, "Error in creating channel for queue "+QUEUE_NAME)

	go consumeMessages(QUEUE_NAME, msgs)

	select {}

}
