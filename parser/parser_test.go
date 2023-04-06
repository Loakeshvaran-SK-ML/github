package parser

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ParserTestSuite struct {
	suite.Suite
}

func (suite *ParserTestSuite) Test_Parser_Invalid_EventType() {
	var wg sync.WaitGroup

	header := map[string]interface{}{
		"X-Github-Event":  []interface{}{"invalid_event_type"},
		"X-Hub-Signature": []interface{}{"abc123"},
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		suite.FailNowf("failed to marshal header to JSON: %v", err.Error())
	}

	data := map[string]interface{}{
		"foo": "bar",
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		suite.FailNowf("failed to marshal data to JSON: %v", err.Error())
	}
	body := string(dataJSON)
	headers := string(headerJSON)
	wg.Add(1)
	err = Parser(&wg, headers, body)
	suite.NotNil(err)
	wg.Wait()
}

func (suite *ParserTestSuite) Test_Parser_InvalidBody() {
	var wg sync.WaitGroup
	wg.Add(1)
	// Test data
	header := "{\"X-Github-Event\": [\"push\"], \"X-Hub-Signature\": [\"signature\"]}"
	body := "invalid body"

	// Call the parser function
	err := Parser(&wg, header, body)

	// Assert that the function returns an error
	suite.Error(err)
	wg.Wait()
}

func (suite *ParserTestSuite) Test_Parser_With_NonJSONData() {
	// Create a wait group for synchronization
	var wg sync.WaitGroup
	wg.Add(1)

	// Define a header and data that are not JSON strings
	header := "this is not a JSON string"
	data := "this is not a JSON string either"

	// Call the Parser function with the non-JSON data
	err := Parser(&wg, header, data)

	// Assert that the Parser function returns an error
	suite.Error(err, "Parser function should return an error with non-JSON header and data")

	// Check the error message to ensure it's related to both header and data
	suite.Contains(err.Error(), "header is not a string")

	fmt.Println("Parser error:", err)
	wg.Wait()
}

func (suite *ParserTestSuite) TestParser() {
	var wg sync.WaitGroup

	header := `{"X-Github-Event": ["push"], "X-Hub-Signature": ["signature123"]}`
	body := `{"head_commit": {"id": "commit123", "timestamp": "2022-04-05T12:34:56Z"}}`

	wg.Add(1)
	err := Parser(&wg, header, body)
	wg.Wait()

	suite.Error(err, "Error occured !")

	// test for error when header is nil
	wg.Add(1)
	err = Parser(&wg, nil, body)
	wg.Wait()

	suite.Error(err, "Expected an error")
	suite.Contains(err.Error(), "Header is nil")

	// test for error when data is nil
	wg.Add(1)
	err = Parser(&wg, header, nil)
	wg.Wait()

	suite.Error(err, "Expected an error")
	suite.Contains(err.Error(), "Data is nil")

	// test for error when header is not a string
	wg.Add(1)
	err = Parser(&wg, 123, body)
	wg.Wait()

	suite.Error(err, "Expected an error")
	suite.Contains(err.Error(), "Header is not a string")

	// test for error when data is not a string
	wg.Add(1)
	err = Parser(&wg, header, 123)
	wg.Wait()

	suite.Error(err, "Expected an error")
	suite.Contains(err.Error(), "Data is not a string")

	// test for error when received empty header message
	wg.Add(1)
	err = Parser(&wg, "", body)
	wg.Wait()

	suite.Error(err, "Expected an error")
	suite.Contains(err.Error(), "Received empty header message")

	// test for error when received empty body message
	wg.Add(1)
	err = Parser(&wg, header, "")
	wg.Wait()

	suite.Error(err, "Expected an error")
	suite.Contains(err.Error(), "Received empty body message")

}

func TestParserTestSuite(t *testing.T) {
	suite.Run(t, new(ParserTestSuite))
}
