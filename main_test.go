package main

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ConsumerTestSuite struct {
	suite.Suite
}

// Set up the test suite
func (suite *ConsumerTestSuite) SetupTest() {
	// Add any necessary set up code here
	
}

// Test the consumeMessages function with an error in the JSON unmarshalling
func (suite *ConsumerTestSuite) Test_ConsumeMessages_With_UnmarshalError() {
	// Create a mock delivery channel
	mockDeliveryChan := make(chan amqp091.Delivery, 1)

	// Send a message with invalid JSON to the channel
	mockDelivery := amqp091.Delivery{Body: []byte(`invalid JSON`)}
	mockDeliveryChan <- mockDelivery

	// Call the consumeMessages function with the mock delivery channel
	go func() {
		err := consumeMessages("test-queue", mockDeliveryChan)
		require.NoError(suite.T(), err)
	}()

	// Expect the message to be skipped and no parser function calls
	mock.AssertExpectationsForObjects(suite.T())
}

func (suite *ConsumerTestSuite) Test_ConsumeMessage_with_Invalideheader() {
	// Create a mock delivery channel
	msg := make(chan amqp091.Delivery)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		bodyData := map[string]interface{}{
			"header": "mock header",
			"body":   "mock body",
		}
		body, _ := json.Marshal(bodyData)
		m := amqp091.Delivery{Body: body}
		msg <- m
		close(msg)
	}()

	// Call the consumeMessages function with the mock delivery channel
	err := consumeMessages("mockQueue", msg)
	if err != nil {
		suite.T().Errorf("consumeMessages returned an error: %v", err)
	}

	wg.Wait()
}

// Testing the header is not string
func (suite *ConsumerTestSuite) Test_ConsumeMessages_header_notstring() {
	// Create a mock messages channel
	mockMsgs := make(chan amqp.Delivery)

	// Create a mock message to send
	header := map[string]string{"key": "value"}
	body := map[string]string{"message": "hello"}
	mockMsgBody, err := json.Marshal(map[string]interface{}{"header": header, "body": body})
	if err != nil {
		suite.T().Fatalf("Failed to marshal mock message: %v", err)
	}
	mockMsg := amqp.Delivery{Body: mockMsgBody}

	// Start the consumeMessages function in a separate goroutine
	go func() {
		err := consumeMessages("testQueue", mockMsgs)
		if err != nil {
			suite.T().Fatalf("consumeMessages failed: %v", err)
		}
	}()

	// Send the mock message to the messages channel
	mockMsgs <- mockMsg

	// Wait for the message to be processed
	time.Sleep(100 * time.Millisecond)
}

// Define the test suite and register the tests
func TestConsumerTestSuite(t *testing.T) {
	suite.Run(t, new(ConsumerTestSuite))
}
